package datatypes

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"strconv"
	"time"
)

var (
	UTC               = time.UTC
	CST               = time.FixedZone("CST", 8*60*60)
	TimestampZero     = Timestamp(time.Time{})
	TimestampUnixZero = Timestamp(time.Unix(0, 0))
)

type MySQLTimestamp = Timestamp

// openapi:strfmt date-time
type Timestamp time.Time

func (dt *Timestamp) DataType(engine string) string {
	return "bigint"
}

func ParseTimestampFromString(s string) (dt Timestamp, err error) {
	var t time.Time
	t, err = time.Parse(time.RFC3339, s)
	dt = Timestamp(t)
	return
}

func ParseTimestampFromStringWithLayout(input, layout string) (Timestamp, error) {
	t, err := time.ParseInLocation(layout, input, CST)
	if err != nil {
		return TimestampUnixZero, err
	}
	return Timestamp(t), nil
}

var _ interface {
	sql.Scanner
	driver.Valuer
} = (*Timestamp)(nil)

func (dt *Timestamp) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return fmt.Errorf("sql.Scan() strfmt.Timestamp from: %#v failed: %s", v, err.Error())
		}
		*dt = Timestamp(time.Unix(n, 0))
	case int64:
		if v < 0 {
			*dt = Timestamp{}
		} else {
			*dt = Timestamp(time.Unix(v, 0))
		}
	case nil:
		*dt = TimestampZero
	default:
		return fmt.Errorf("cannot sql.Scan() strfmt.Timestamp from: %#v", v)
	}
	return nil
}

func (dt Timestamp) Value() (driver.Value, error) {
	s := (time.Time)(dt).Unix()
	if s < 0 {
		s = 0
	}
	return s, nil
}

func (dt Timestamp) String() string {
	if dt.IsZero() {
		return ""
	}
	return time.Time(dt).In(CST).Format(time.RFC3339)
}

func (dt Timestamp) Format(layout string) string {
	return time.Time(dt).In(CST).Format(layout)
}

var _ interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
} = (*Timestamp)(nil)

func (dt Timestamp) MarshalText() ([]byte, error) {
	return []byte(dt.String()), nil
}

func (dt *Timestamp) UnmarshalText(data []byte) (err error) {
	str := string(data)
	if len(str) == 0 || str == "0" {
		return nil
	}
	*dt, err = ParseTimestampFromString(str)
	return
}

func (dt Timestamp) Unix() int64 {
	return time.Time(dt).Unix()
}

func (dt Timestamp) IsZero() bool {
	unix := dt.Unix()
	return unix == 0 || unix == TimestampZero.Unix()
}

func (dt Timestamp) In(loc *time.Location) Timestamp {
	return Timestamp(time.Time(dt).In(loc))
}
