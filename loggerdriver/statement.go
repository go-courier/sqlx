package loggerdriver

import (
	"database/sql/driver"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

var _ driver.Stmt = (*loggerStmt)(nil)

type loggerStmt struct {
	logger *logrus.Logger
	cfg    *mysql.Config
	query  string
	stmt   driver.Stmt
}

func (s *loggerStmt) Close() error {
	if err := s.stmt.Close(); err != nil {
		s.logger.Errorf("failed to close statement: %s", err)
		return err
	}
	return nil
}

var DuplicateEntryErrNumber uint16 = 1062

func startTimer() func() time.Duration {
	startTime := time.Now()
	return func() time.Duration {
		return time.Now().Sub(startTime)
	}
}

func (s *loggerStmt) Exec(args []driver.Value) (driver.Result, error) {
	cost := startTimer()

	if len(args) != 0 {
		sqlForLog, err := s.interpolateParams(s.query, args)
		if err != nil {
			s.logger.Warnf("failed exec %s: %s", err, color.RedString(s.query))
			return nil, err
		}
		s.query = sqlForLog
	}

	result, err := s.stmt.Exec(args)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); !ok {
			s.logger.Errorf("failed exec %s: %s", err, color.RedString(s.query))
		} else if mysqlErr.Number == DuplicateEntryErrNumber {
			s.logger.Warnf("failed exec %s: %s", err, color.RedString(s.query))
		} else {
			s.logger.Errorf("failed exec %s: %s", err, color.RedString(s.query))
		}
		return nil, err
	}

	s.logger.WithField("cost", cost().String()).Debugf(color.YellowString(s.query))
	return result, nil
}

func (s *loggerStmt) Query(args []driver.Value) (driver.Rows, error) {
	cost := startTimer()

	if len(args) != 0 {
		sqlForLog, err := s.interpolateParams(s.query, args)
		if err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); !ok {
				s.logger.Errorf("failed exec %s: %s", err, color.RedString(s.query))
			} else {
				s.logger.Warnf("failed exec %s: %s", mysqlErr, color.RedString(s.query))
			}
			return nil, err
		}
		s.query = sqlForLog
	}

	rows, err := s.stmt.Query(args)
	if err != nil {
		s.logger.Warnf("failed query %s: %s", err, color.RedString(s.query))
		return nil, err
	}

	s.logger.WithField("cost", cost().String()).Debugf(color.GreenString(s.query))
	return rows, nil
}

func (s *loggerStmt) NumInput() int {
	i := s.stmt.NumInput()
	return i
}

func (s *loggerStmt) interpolateParams(query string, args []driver.Value) (string, error) {
	if strings.Count(query, "?") != len(args) {
		return "", driver.ErrSkip
	}

	buf := make([]byte, 0)
	buf = buf[:0]
	argPos := 0

	for i := 0; i < len(query); i++ {
		q := strings.IndexByte(query[i:], '?')
		if q == -1 {
			buf = append(buf, query[i:]...)
			break
		}
		buf = append(buf, query[i:i+q]...)
		i += q

		arg := args[argPos]
		argPos++

		if arg == nil {
			buf = append(buf, "NULL"...)
			continue
		}

		switch v := arg.(type) {
		case int64:
			buf = strconv.AppendInt(buf, v, 10)
		case float64:
			buf = strconv.AppendFloat(buf, v, 'g', -1, 64)
		case bool:
			if v {
				buf = append(buf, '1')
			} else {
				buf = append(buf, '0')
			}
		case time.Time:
			if v.IsZero() {
				buf = append(buf, "'0000-00-00'"...)
			} else {
				v := v.In(s.cfg.Loc)
				v = v.Add(time.Nanosecond * 500) // Write round under microsecond
				year := v.Year()
				year100 := year / 100
				year1 := year % 100
				month := v.Month()
				day := v.Day()
				hour := v.Hour()
				minute := v.Minute()
				second := v.Second()
				micro := v.Nanosecond() / 1000

				buf = append(buf, []byte{
					'\'',
					digits10[year100], digits01[year100],
					digits10[year1], digits01[year1],
					'-',
					digits10[month], digits01[month],
					'-',
					digits10[day], digits01[day],
					' ',
					digits10[hour], digits01[hour],
					':',
					digits10[minute], digits01[minute],
					':',
					digits10[second], digits01[second],
				}...)

				if micro != 0 {
					micro10000 := micro / 10000
					micro100 := micro / 100 % 100
					micro1 := micro % 100
					buf = append(buf, []byte{
						'.',
						digits10[micro10000], digits01[micro10000],
						digits10[micro100], digits01[micro100],
						digits10[micro1], digits01[micro1],
					}...)
				}
				buf = append(buf, '\'')
			}
		case []byte:
			if v == nil {
				buf = append(buf, "NULL"...)
			} else {
				buf = append(buf, "_binary'"...)
				buf = escapeBytesBackslash(buf, v)
				buf = append(buf, '\'')
			}
		case string:
			buf = append(buf, '\'')
			buf = escapeBytesBackslash(buf, []byte(v))
			buf = append(buf, '\'')
		default:
			return "", driver.ErrSkip
		}

		if len(buf)+4 > s.cfg.MaxAllowedPacket {
			return "", driver.ErrSkip
		}
	}
	if argPos != len(args) {
		return "", driver.ErrSkip
	}
	return string(buf), nil
}

// copy from mysql driver

const digits01 = "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
const digits10 = "0000000000111111111122222222223333333333444444444455555555556666666666777777777788888888889999999999"

func escapeBytesBackslash(buf, v []byte) []byte {
	pos := len(buf)
	buf = reserveBuffer(buf, len(v)*2)

	for _, c := range v {
		switch c {
		case '\x00':
			buf[pos] = '\\'
			buf[pos+1] = '0'
			pos += 2
		case '\n':
			buf[pos] = '\\'
			buf[pos+1] = 'n'
			pos += 2
		case '\r':
			buf[pos] = '\\'
			buf[pos+1] = 'r'
			pos += 2
		case '\x1a':
			buf[pos] = '\\'
			buf[pos+1] = 'Z'
			pos += 2
		case '\'':
			buf[pos] = '\\'
			buf[pos+1] = '\''
			pos += 2
		case '"':
			buf[pos] = '\\'
			buf[pos+1] = '"'
			pos += 2
		case '\\':
			buf[pos] = '\\'
			buf[pos+1] = '\\'
			pos += 2
		default:
			buf[pos] = c
			pos++
		}
	}

	return buf[:pos]
}

func reserveBuffer(buf []byte, appendSize int) []byte {
	newSize := len(buf) + appendSize
	if cap(buf) < newSize {
		// Grow buffer exponentially
		newBuf := make([]byte, len(buf)*2+appendSize)
		copy(newBuf, buf)
		buf = newBuf
	}
	return buf[:newSize]
}
