package mysqlconnector

import (
	"database/sql/driver"
	"errors"
	"strconv"
	"strings"
	"time"
)

func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			// TODO: support the use of Named Parameters #561
			return nil, errors.New("mysql: driver does not support the use of Named Parameters")
		}
		dargs[n] = param.Value
	}
	return dargs, nil
}

func interpolateParams(query string, args []driver.Value, loc *time.Location, maxAllowedPacket int) (string, error) {
	if strings.Count(query, "?") != len(args) {
		return "", driver.ErrSkip
	}

	buf := make([]byte, 0)
	buf = buf[:0]
	argPos := 0

	data := []byte(query)

	for i := range data {
		q := query[i]
		switch q {
		case '?':
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
					v := v.In(loc)
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

		case '\n':
			buf = append(buf, ' ')
		default:
			buf = append(buf, q)
		}

		if len(buf)+4 > maxAllowedPacket {
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
