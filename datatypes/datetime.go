package datatypes

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"time"
)

var (
	DatetimeZero = Datetime(time.Time{})
)

type MySQLDatetime = Datetime

// openapi:strfmt date-time
type Datetime time.Time

func (Datetime) DataType(e string) string {
	return "timestamp"
}

func ParseDatetimeFromString(s string) (dt Datetime, err error) {
	var t time.Time
	t, err = time.Parse(time.RFC3339, s)
	dt = Datetime(t)
	return
}

func ParseDatetimeFromStringWithFormatterInCST(s, formatter string) (dt Datetime, err error) {
	var t time.Time
	t, err = time.ParseInLocation(formatter, s, CST)
	dt = Datetime(t)
	return
}

var _ interface {
	sql.Scanner
	driver.Valuer
} = (*Datetime)(nil)

func (dt *Datetime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		*dt = Datetime(time.Unix(v.Unix(), 0))
	case nil:
		*dt = DatetimeZero
	default:
		return fmt.Errorf("cannot sql.Scan() strfmt.Datetime from: %#v", v)
	}
	return nil
}

func (dt Datetime) Value() (driver.Value, error) {
	return time.Unix(dt.Unix(), 0), nil
}

func (dt Datetime) String() string {
	if dt.IsZero() {
		return ""
	}
	return time.Time(dt).In(CST).Format(time.RFC3339)
}

func (dt Datetime) Format(layout string) string {
	return time.Time(dt).In(CST).Format(layout)
}

var _ interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
} = (*Datetime)(nil)

func (dt Datetime) MarshalText() ([]byte, error) {
	return []byte(dt.String()), nil
}

func (dt *Datetime) UnmarshalText(data []byte) (err error) {
	str := string(data)
	if len(str) == 0 || str == "0" {
		return nil
	}
	*dt, err = ParseDatetimeFromString(str)
	return
}

func (dt Datetime) Unix() int64 {
	return time.Time(dt).Unix()
}

func (dt Datetime) IsZero() bool {
	unix := dt.Unix()
	return unix == 0 || unix == DatetimeZero.Unix()
}

func (dt Datetime) In(loc *time.Location) Datetime {
	return Datetime(time.Time(dt).In(loc))
}
