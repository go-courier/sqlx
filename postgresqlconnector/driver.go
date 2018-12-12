package postgresqlconnector

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var _ interface {
	driver.Driver
} = (*PostgreSQLLoggingDriver)(nil)

type PostgreSQLLoggingDriver struct {
	Logger *logrus.Logger
	Driver *pq.Driver
}

func FromConfigString(s string) PostgreSQLOpts {
	opts := PostgreSQLOpts{}
	for _, kv := range strings.Split(s, " ") {
		kvs := strings.Split(kv, "=")
		if len(kvs) > 1 {
			opts[kvs[0]] = kvs[1]
		}
	}
	return opts
}

type PostgreSQLOpts map[string]string

func (opts PostgreSQLOpts) String() string {
	buf := bytes.NewBuffer(nil)

	kvs := make([]string, 0)
	for k := range opts {
		kvs = append(kvs, k)
	}
	sort.Strings(kvs)

	for i, k := range kvs {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(opts[k])
	}

	return buf.String()
}

func (d *PostgreSQLLoggingDriver) Open(dsn string) (driver.Conn, error) {
	conf, err := pq.ParseURL(dsn)
	if err != nil {
		panic(err)
	}
	opts := FromConfigString(conf)
	if pass, ok := opts["password"]; ok {
		opts["password"] = strings.Repeat("*", len(pass))
	}

	conn, err := d.Driver.Open(conf)
	if err != nil {
		d.Logger.Errorf("failed to open connection: %s %s", opts, err)
		return nil, err
	}
	d.Logger.Debugf(color.YellowString("connected %s", opts))
	return &loggerConn{Conn: conn, cfg: opts, logger: d.Logger}, nil
}

var _ interface {
	driver.Conn
	driver.ExecerContext
	driver.QueryerContext
} = (*loggerConn)(nil)

type loggerConn struct {
	logger *logrus.Logger
	cfg    PostgreSQLOpts
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

func (c *loggerConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	cost := startTimer()

	defer func() {
		query = interpolateParams(query, args)

		if err != nil {
			if pgErr, ok := err.(*pq.Error); !ok {
				c.logger.Errorf("failed query %s: %s", err, color.RedString(query))
			} else {
				c.logger.Warnf("failed query %s: %s", pgErr, color.RedString(query))
			}
		} else {
			c.logger.WithField("cost", cost().String()).Debugf(color.YellowString(query))
		}
	}()

	rows, err = c.Conn.(driver.QueryerContext).QueryContext(ctx, query, args)
	return
}

func (c *loggerConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (result driver.Result, err error) {
	cost := startTimer()

	defer func() {
		query = interpolateParams(query, args)

		if err != nil {
			if pgError, ok := err.(*pq.Error); !ok {
				c.logger.Errorf("failed exec %s: %s", err, color.RedString(query))
			} else if pgError.Code == "23505" {
				c.logger.Warnf("failed exec %s: %s", err, color.RedString(query))
			} else {
				c.logger.Errorf("failed exec %s: %s", pgError, color.RedString(query))
			}
			return
		}

		c.logger.WithField("cost", cost().String()).Debugf(color.YellowString(query))
	}()

	result, err = c.Conn.(driver.ExecerContext).ExecContext(ctx, query, args)
	return
}

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
