package mysqlconnector

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
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

	d.Logger.Debugf(color.YellowString("connected %s", cfg.FormatDSN()))
	return &loggerConn{Conn: conn, cfg: cfg, logger: d.Logger}, nil
}

var _ interface {
	driver.Conn
	driver.ExecerContext
	driver.QueryerContext
} = (*loggerConn)(nil)

type loggerConn struct {
	logger *logrus.Logger
	cfg    *mysql.Config
	driver.Conn
}

func (c *loggerConn) Begin() (driver.Tx, error) {
	c.logger.Debugf(color.YellowString("=========== Beginning Transaction ==========="))
	tx, err := c.Conn.Begin()
	if err != nil {
		c.logger.Errorf("failed to begin transaction: %s", err)
		return nil, err
	}
	return &loggingTx{tx: tx, logger: c.logger}, nil
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

func (s *loggerConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	cost := startTimer()

	defer func() {
		query = s.interpolateParams(query, args)

		if err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); !ok {
				s.logger.Errorf("failed query %s: %s", err, color.RedString(query))
			} else {
				s.logger.Warnf("failed query %s: %s", mysqlErr, color.RedString(query))
			}
		} else {
			s.logger.WithField("cost", cost().String()).Debugf(color.YellowString(query))
		}
	}()

	rows, err = s.Conn.(driver.QueryerContext).QueryContext(ctx, query, args)
	return
}

func (s *loggerConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (result driver.Result, err error) {
	cost := startTimer()

	defer func() {
		query = s.interpolateParams(query, args)

		if err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); !ok {
				s.logger.Errorf("failed exec %s: %s", err, color.RedString(query))
			} else if mysqlErr.Number == DuplicateEntryErrNumber {
				s.logger.Warnf("failed exec %s: %s", err, color.RedString(query))
			} else {
				s.logger.Errorf("failed exec %s: %s", mysqlErr, color.RedString(query))
			}
			return
		}

		s.logger.WithField("cost", cost().String()).Debugf(color.YellowString(query))
	}()

	result, err = s.Conn.(driver.ExecerContext).ExecContext(ctx, query, args)
	return
}

func (s *loggerConn) interpolateParams(query string, args []driver.NamedValue) string {
	if len(args) == 0 {
		return query
	}
	argValues, err := namedValueToValue(args)
	if err != nil {
		return query
	}
	sqlForLog, err := interpolateParams(query, argValues, s.cfg.Loc, s.cfg.MaxAllowedPacket)
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
	logger *logrus.Logger
	tx     driver.Tx
}

func (tx *loggingTx) Commit() error {
	if err := tx.tx.Commit(); err != nil {
		tx.logger.Debugf("failed to commit transaction: %s", err)
		return err
	}
	tx.logger.Debugf(color.YellowString("=========== Committed Transaction ==========="))
	return nil
}

func (tx *loggingTx) Rollback() error {
	if err := tx.tx.Rollback(); err != nil {
		tx.logger.Debugf("failed to rollback transaction: %s", err)
		return err
	}
	tx.logger.Debugf("=========== Rollback Transaction ===========")
	return nil
}
