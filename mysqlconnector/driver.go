package mysqlconnector

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/go-courier/logr"
	"github.com/pkg/errors"

	"github.com/go-sql-driver/mysql"
)

func init() {
	_ = mysql.SetLogger(&logger{})
}

type logger struct{}

func (l *logger) Print(args ...interface{}) {
}

var _ interface {
	driver.Driver
} = (*MySqlLoggingDriver)(nil)

type MySqlLoggingDriver struct {
	driver mysql.MySQLDriver
}

func (d *MySqlLoggingDriver) Open(dsn string) (driver.Conn, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	cfg.Passwd = strings.Repeat("*", len(cfg.Passwd))

	conn, err := d.driver.Open(dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open connection: %s", cfg.FormatDSN())
	}
	return &loggerConn{Conn: conn, cfg: cfg}, nil
}

func (d *MySqlLoggingDriver) Driver() driver.Driver {
	return d
}

var _ interface {
	driver.ConnBeginTx
	driver.ExecerContext
	driver.QueryerContext
} = (*loggerConn)(nil)

type loggerConn struct {
	cfg *mysql.Config
	driver.Conn
}

func (c *loggerConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	logger := logr.FromContext(ctx)

	logger.Debug("=========== Beginning Transaction ===========")
	tx, err := c.Conn.(driver.ConnBeginTx).BeginTx(ctx, opts)
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to begin transaction"))
		return nil, err
	}
	return &loggingTx{Tx: tx, logger: logger}, nil
}

func (c *loggerConn) Prepare(query string) (driver.Stmt, error) {
	panic(fmt.Errorf("don't use Prepare"))
}

func (c *loggerConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	cost := startTimer()
	newCtx, logger := logr.Start(ctx, "Query")

	defer func() {
		q := c.interpolateParams(query, args)

		if err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); !ok {
				logger.Error(errors.Wrapf(err, "query failed: %s", q))
			} else {
				logger.Warn(errors.Wrapf(mysqlErr, "query failed: %s", q))
			}
		} else {
			logger.WithValues("cost", cost().String()).Debug(q.String())
		}

		logger.End()
	}()

	rows, err = c.Conn.(driver.QueryerContext).QueryContext(newCtx, query, args)
	return
}

func (c *loggerConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (result driver.Result, err error) {
	cost := startTimer()
	newCtx, logger := logr.Start(ctx, "Query")

	defer func() {
		q := c.interpolateParams(query, args)

		if err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); !ok {
				logger.Error(errors.Wrapf(err, "exec failed: %s", q))
			} else if mysqlErr.Number == DuplicateEntryErrNumber {
				logger.Error(errors.Wrapf(err, "exec failed: %s", q))
			} else {
				logger.Warn(errors.Wrapf(mysqlErr, "exec failed: %s", q))
			}
		} else {
			logger.WithValues("cost", cost().String()).Debug(q.String())
		}

		logger.End()
	}()

	result, err = c.Conn.(driver.ExecerContext).ExecContext(newCtx, query, args)
	return
}

func (c *loggerConn) interpolateParams(query string, args []driver.NamedValue) fmt.Stringer {
	return &SqlPrinter{query, args, c.cfg}
}

type SqlPrinter struct {
	query string
	args  []driver.NamedValue
	cfg   *mysql.Config
}

func (p *SqlPrinter) String() string {
	if len(p.args) == 0 {
		return p.query
	}
	argValues, err := namedValueToValue(p.args)
	if err != nil {
		return p.query
	}
	sqlForLog, err := interpolateParams(p.query, argValues, p.cfg.Loc, p.cfg.MaxAllowedPacket)
	if err != nil {
		return p.query
	}

	return sqlForLog
}

var DuplicateEntryErrNumber uint16 = 1062

func startTimer() func() time.Duration {
	startTime := time.Now()
	return func() time.Duration {
		return time.Since(startTime)
	}
}

type loggingTx struct {
	logger logr.Logger
	driver.Tx
}

func (tx *loggingTx) Commit() error {
	if err := tx.Tx.Commit(); err != nil {
		tx.logger.Debug("failed to commit transaction: %s", err)
		return err
	}
	tx.logger.Debug("=========== Committed Transaction ===========")
	return nil
}

func (tx *loggingTx) Rollback() error {
	if err := tx.Tx.Rollback(); err != nil {
		tx.logger.Debug("failed to rollback transaction: %s", err)
		return err
	}
	tx.logger.Debug("=========== Rollback Transaction ===========")
	return nil
}
