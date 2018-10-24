package loggermysql

import (
	"database/sql"
	"database/sql/driver"
	"strings"

	"github.com/fatih/color"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

func init() {
	sql.Register("logger:mysql", &LoggingDriver{Driver: &mysql.MySQLDriver{}, Logger: logrus.StandardLogger()})
	mysql.SetLogger(&logger{})
}

type logger struct{}

func (l *logger) Print(args ...interface{}) {
}

var _ driver.Driver = (*LoggingDriver)(nil)

type LoggingDriver struct {
	Logger *logrus.Logger
	Driver *mysql.MySQLDriver
}

func (d *LoggingDriver) Open(name string) (driver.Conn, error) {
	cfg, err := mysql.ParseDSN(name)
	if err != nil {
		panic(err)
	}
	cfg.Passwd = strings.Repeat("*", len(cfg.Passwd))

	conn, err := d.Driver.Open(name)
	if err != nil {
		d.Logger.Errorf("failed to open connection: %s %s", cfg.FormatDSN(), err)
		return nil, err
	}

	d.Logger.Debugf(color.YellowString("connected %s", cfg.FormatDSN()))

	return &loggerConn{cfg: cfg, conn: conn, logger: d.Logger}, nil
}
