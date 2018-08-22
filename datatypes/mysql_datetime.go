package datatypes

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"time"
)

var (
	MySQLDatetimeZero = MySQLDatetime(time.Time{})
)

// swagger:strfmt date-time
type MySQLDatetime time.Time

func ParseMySQLDatetimeFromString(s string) (dt MySQLDatetime, err error) {
	var t time.Time
	t, err = time.Parse(time.RFC3339, s)
	dt = MySQLDatetime(t)
	return
}

func ParseMySQLDatetimeFromStringWithFormatterInCST(s, formatter string) (dt MySQLDatetime, err error) {
	var t time.Time
	t, err = time.ParseInLocation(formatter, s, CST)
	dt = MySQLDatetime(t)
	return
}

var _ interface {
	sql.Scanner
	driver.Valuer
} = (*MySQLDatetime)(nil)

func (dt *MySQLDatetime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		*dt = MySQLDatetime(time.Unix(v.Unix(), 0))
	case nil:
		*dt = MySQLDatetimeZero
	default:
		return fmt.Errorf("cannot sql.Scan() strfmt.MySQLDatetime from: %#v", v)
	}
	return nil
}

func (dt MySQLDatetime) Value() (driver.Value, error) {
	return time.Unix(dt.Unix(), 0), nil
}

func (dt MySQLDatetime) String() string {
	return time.Time(dt).In(CST).Format(time.RFC3339)
}

func (dt MySQLDatetime) Format(layout string) string {
	return time.Time(dt).In(CST).Format(layout)
}

var _ interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
} = (*MySQLDatetime)(nil)

func (dt MySQLDatetime) MarshalText() ([]byte, error) {
	if dt.IsZero() {
		return []byte(""), nil
	}
	str := dt.String()
	return []byte(str), nil
}

func (dt *MySQLDatetime) UnmarshalText(data []byte) (err error) {
	str := string(data)
	if len(str) == 0 || str == "0" {
		str = MySQLDatetimeZero.String()
	}
	*dt, err = ParseMySQLDatetimeFromString(str)
	return
}

func (dt MySQLDatetime) Unix() int64 {
	return time.Time(dt).Unix()
}

func (dt MySQLDatetime) IsZero() bool {
	unix := dt.Unix()
	return unix == 0 || unix == MySQLDatetimeZero.Unix()
}

func (dt MySQLDatetime) In(loc *time.Location) MySQLDatetime {
	return MySQLDatetime(time.Time(dt).In(loc))
}
