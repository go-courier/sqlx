package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestColumns(t *testing.T) {
	tt := require.New(t)

	columns := Columns{}

	tt.Equal(0, columns.Len())

	{
		col := columns.AutoIncrement()
		tt.Nil(col)
	}

	{
		columns.Add(Col(nil, "F_id").Field("ID").Type("bigint(64) unsigned NOT NULL AUTO_INCREMENT"))

		col := columns.AutoIncrement()
		tt.NotNil(col)
		tt.Equal("F_id", col.Name)
	}

	tt.Nil(columns.F("ID2"))
	tt.Equal(0, columns.Fields("ID2").Len())
	tt.Equal(1, columns.Fields().Len())
	tt.Len(columns.Fields("ID2").List(), 0)
	tt.Equal(1, columns.Cols("F_id").Len())
	tt.Equal(1, columns.Cols().Len())
	tt.Len(columns.Cols("F_id").List(), 1)
	tt.Equal([]string{"ID"}, columns.Cols("F_id").FieldNames())
}
