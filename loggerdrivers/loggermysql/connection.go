package loggermysql

import (
	"database/sql/driver"

	"github.com/fatih/color"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

var _ interface {
	driver.Conn
} = (*loggerConn)(nil)

type loggerConn struct {
	logger *logrus.Logger
	cfg    *mysql.Config
	conn   driver.Conn
}

func (c *loggerConn) Begin() (driver.Tx, error) {
	c.logger.Debugf(color.YellowString("=========== Beginning Transaction ==========="))
	tx, err := c.conn.Begin()
	if err != nil {
		c.logger.Errorf("failed to begin transaction: %s", err)
		return nil, err
	}
	return &loggingTx{tx: tx, logger: c.logger}, nil
}

func (c *loggerConn) Close() error {
	if err := c.conn.Close(); err != nil {
		c.logger.Errorf("failed to close connection: %s", err)
		return err
	}
	return nil
}

func (c *loggerConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.conn.Prepare(query)
	if err != nil {
		c.logger.Errorf("failed to prepare query: %s, err: %s", query, err)
		return nil, err
	}
	return &loggerStmt{cfg: c.cfg, query: query, stmt: stmt, logger: c.logger}, nil
}
