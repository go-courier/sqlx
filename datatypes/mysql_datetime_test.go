package datatypes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDatetime(t *testing.T) {
	tt := require.New(t)

	t0, _ := time.Parse(time.RFC3339, "2017-03-27T23:58:59+08:00")
	dt := MySQLDatetime(t0)
	tt.Equal("2017-03-27T23:58:59+08:00", dt.String())
	tt.Equal("2017-03-27T23:58:59+08:00", dt.Format(time.RFC3339))
	tt.Equal(int64(1490630339), dt.Unix())

	{
		dateString, err := dt.MarshalText()
		tt.NoError(err)
		tt.Equal("2017-03-27T23:58:59+08:00", string(dateString))

		dt2 := MySQLDatetimeZero
		tt.True(dt2.IsZero())
		err = dt2.UnmarshalText(dateString)
		tt.NoError(err)
		tt.Equal(dt, dt2)
		tt.False(dt2.IsZero())
	}

	{
		value, err := dt.Value()
		tt.NoError(err)
		tt.Equal("2017-03-27T23:58:59+08:00", value.(time.Time).In(CST).Format(time.RFC3339))

		dt2 := MySQLDatetimeZero
		tt.True(dt2.IsZero())
		err = dt2.Scan(value)
		tt.NoError(err)
		tt.Equal(dt.In(CST), dt2.In(CST))
		tt.False(dt2.IsZero())
	}

	{
		dt3 := MySQLTimestampZero
		err := dt3.UnmarshalText([]byte(""))
		tt.NoError(err)
	}
}
