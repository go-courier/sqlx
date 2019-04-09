package mysqlconnector

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

func init() {
	mysql.SetLogger(&logger{})
}

type logger struct{}

func (l *logger) Print(args ...interface{}) {
}

var _ driver.Driver = (*MySqlLoggingDriver)(nil)

type MySqlLoggingDriver struct {
	Logger *logrus.Logger
	Driver *mysql.MySQLDriver
}

func (d *MySqlLoggingDriver) Open(dsn string) (driver.Conn, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		panic(err)
	}
	cfg.Passwd = strings.Repeat("*", len(cfg.Passwd))
	conn, err := d.Driver.Open(dsn)
	if err != nil {
		d.Logger.Errorf("failed to open connection: %s %s", cfg.FormatDSN(), err)
		return nil, err
	}

	d.Logger.Debugf("connected %s", cfg.FormatDSN())
	return &loggerConn{Conn: conn, cfg: cfg, logger: d.Logger.WithField("driver", "mysql")}, nil
}

var _ interface {
	driver.ConnBeginTx
	driver.ExecerContext
	driver.QueryerContext
} = (*loggerConn)(nil)

type loggerConn struct {
	logger *logrus.Entry
	cfg    *mysql.Config
	driver.Conn
}

func (c *loggerConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	logger := c.logger.WithContext(ctx)

	logger.Debug("=========== Beginning Transaction ===========")
	tx, err := c.Conn.(driver.ConnBeginTx).BeginTx(ctx, opts)
	if err != nil {
		logger.Errorf("failed to begin transaction: %s", err)
		return nil, err
	}
	return &loggingTx{Tx: tx, logger: logger}, nil
}

func (c *loggerConn) Close() error {
	if err := c.Conn.Close(); err != nil {
		c.logger.Errorf("failed to close connection: %s", err)
		return err
	}
	return nil
}

func (c *loggerConn) Prepare(query string) (driver.Stmt, error) {
	panic(fmt.Errorf("don't use Prepare"))
}

func (c *loggerConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	cost := startTimer()
	logger := c.logger.WithContext(ctx)

	defer func() {
		query = c.interpolateParams(query, args)

		if err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); !ok {
				logger.Errorf("failed query %s: %s", err, query)
			} else {
				logger.Warnf("failed query %s: %s", mysqlErr, query)
			}
		} else {
			logger.WithField("cost", cost().String()).Debug(query)
		}
	}()

	rows, err = c.Conn.(driver.QueryerContext).QueryContext(ctx, query, args)
	return
}

func (c *loggerConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (result driver.Result, err error) {
	cost := startTimer()
	logger := c.logger.WithContext(ctx)

	defer func() {
		query = c.interpolateParams(query, args)

		if err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); !ok {
				logger.Errorf("failed exec %s: %s", err, query)
			} else if mysqlErr.Number == DuplicateEntryErrNumber {
				logger.Warnf("failed exec %s: %s", err, query)
			} else {
				logger.Errorf("failed exec %s: %s", mysqlErr, query)
			}
			return
		}

		logger.WithField("cost", cost().String()).Debug(query)
	}()

	result, err = c.Conn.(driver.ExecerContext).ExecContext(ctx, query, args)
	return
}

func (c *loggerConn) interpolateParams(query string, args []driver.NamedValue) string {
	if len(args) == 0 {
		return query
	}
	argValues, err := namedValueToValue(args)
	if err != nil {
		return query
	}
	sqlForLog, err := interpolateParams(query, argValues, c.cfg.Loc, c.cfg.MaxAllowedPacket)
	if err != nil {
		return query
	}
	return sqlForLog
}

var DuplicateEntryErrNumber uint16 = 1062

func startTimer() func() time.Duration {
	startTime := time.Now()
	return func() time.Duration {
		return time.Now().Sub(startTime)
	}
}

type loggingTx struct {
	logger *logrus.Entry
	driver.Tx
}

func (tx *loggingTx) Commit() error {
	if err := tx.Tx.Commit(); err != nil {
		tx.logger.Debugf("failed to commit transaction: %s", err)
		return err
	}
	tx.logger.Debug("=========== Committed Transaction ===========")
	return nil
}

func (tx *loggingTx) Rollback() error {
	if err := tx.Tx.Rollback(); err != nil {
		tx.logger.Debugf("failed to rollback transaction: %s", err)
		return err
	}
	tx.logger.Debug("=========== Rollback Transaction ===========")
	return nil
}
