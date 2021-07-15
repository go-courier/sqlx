package postgresqlconnector

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-courier/sqlx/v2"

	"github.com/go-courier/logr"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

var _ interface {
	driver.Driver
} = (*PostgreSQLLoggingDriver)(nil)

type PostgreSQLLoggingDriver struct {
	driver pq.Driver
}

func (d *PostgreSQLLoggingDriver) Open(dsn string) (driver.Conn, error) {
	config, err := pq.ParseURL(dsn)
	if err != nil {
		return nil, err
	}

	opts := FromConfigString(config)
	if pass, ok := opts["password"]; ok {
		opts["password"] = strings.Repeat("*", len(pass))
	}

	conn, err := d.driver.Open(config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open connection: %s", opts)
	}

	return &loggerConn{Conn: conn, cfg: opts}, nil
}

var _ interface {
	driver.ConnBeginTx
	driver.ExecerContext
	driver.QueryerContext
} = (*loggerConn)(nil)

type loggerConn struct {
	cfg PostgreSQLOpts
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
	return &loggingTx{tx: tx, logger: logger}, nil
}

func (c *loggerConn) Close() error {
	if err := c.Conn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *loggerConn) Prepare(query string) (driver.Stmt, error) {
	panic(fmt.Errorf("don't use Prepare"))
}

func (c *loggerConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	newCtx, logger := logr.Start(ctx, "Query")
	cost := startTimer()

	defer func() {
		q := interpolateParams(query, args)

		if err != nil {
			if pgErr, ok := sqlx.UnwrapAll(err).(*pq.Error); !ok {
				logger.Error(errors.Wrapf(pgErr, "query failed: %s", q))
			} else {
				logger.Warn(errors.Wrapf(pgErr, "query failed: %s", q))
			}
		} else {
			logger.WithValues("cost", cost().String()).Debug("%s", q)
		}

		logger.End()
	}()

	rows, err = c.Conn.(driver.QueryerContext).QueryContext(newCtx, replaceValueHolder(query), args)
	return
}

func (c *loggerConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (result driver.Result, err error) {
	cost := startTimer()
	newCtx, logger := logr.Start(ctx, "Exec")

	defer func() {
		q := interpolateParams(query, args)

		if err != nil {
			if pgError, ok := sqlx.UnwrapAll(err).(*pq.Error); !ok {
				logger.Error(errors.Wrapf(err, "exec failed: %s", q))
			} else if pgError.Code == "23505" {
				logger.Warn(errors.Wrapf(pgError, "exec failed: %s", q))
			} else {
				logger.Error(errors.Wrapf(pgError, "exec failed: %s", q))
			}
			return
		}

		logger.WithValues("cost", cost().String()).Debug(q.String())

		logger.End()
	}()

	result, err = c.Conn.(driver.ExecerContext).ExecContext(newCtx, replaceValueHolder(query), args)
	return
}

func replaceValueHolder(query string) string {
	index := 0
	data := []byte(query)

	e := bytes.NewBufferString("")

	for i := range data {
		c := data[i]
		switch c {
		case '?':
			e.WriteByte('$')
			e.WriteString(strconv.FormatInt(int64(index+1), 10))
			index++
		default:
			e.WriteByte(c)
		}
	}

	return e.String()
}

func startTimer() func() time.Duration {
	startTime := time.Now()
	return func() time.Duration {
		return time.Since(startTime)
	}
}

type loggingTx struct {
	logger logr.Logger
	tx     driver.Tx
}

func (tx *loggingTx) Commit() error {
	if err := tx.tx.Commit(); err != nil {
		tx.logger.Debug("failed to commit transaction: %s", err)
		return err
	}
	tx.logger.Debug("=========== Committed Transaction ===========")
	return nil
}

func (tx *loggingTx) Rollback() error {
	if err := tx.tx.Rollback(); err != nil {
		tx.logger.Debug("failed to rollback transaction: %s", err)
		return err
	}
	tx.logger.Debug("=========== Rollback Transaction ===========")
	return nil
}
