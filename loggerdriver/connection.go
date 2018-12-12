package loggerdriver

import (
	"context"
	"database/sql/driver"
	"github.com/fatih/color"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

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

func (s *loggerConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	cost := startTimer()

	defer func() {
		if err != nil {
			argValues, e := namedValueToValue(args)
			if e == nil {
				sqlForLog, e := interpolateParams(s.cfg, query, argValues)
				if e == nil {
					if mysqlErr, ok := err.(*mysql.MySQLError); !ok {
						s.logger.Errorf("failed query %s: %s", err, color.RedString(sqlForLog))
					} else {
						s.logger.Warnf("failed query %s: %s", mysqlErr, color.RedString(sqlForLog))
					}
				}
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
		if err != nil {
			argValues, e := namedValueToValue(args)
			if e == nil {
				sqlForLog, e := interpolateParams(s.cfg, query, argValues)
				if e == nil {
					if mysqlErr, ok := err.(*mysql.MySQLError); !ok {
						s.logger.Errorf("failed exec %s: %s", err, color.RedString(sqlForLog))
					} else {
						s.logger.Warnf("failed exec %s: %s", mysqlErr, color.RedString(sqlForLog))
					}
				}
			}
		} else {
			s.logger.WithField("cost", cost().String()).Debugf(color.YellowString(query))
		}
	}()

	result, err = s.Conn.(driver.ExecerContext).ExecContext(ctx, query, args)
	return
}
