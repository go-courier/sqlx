package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestColumns(t *testing.T) {
	columns := Columns{}

	require.Equal(t, 0, columns.Len())

	{
		col := columns.AutoIncrement()
		require.Nil(t, col)
	}

	{
		columns.Add(
			Col("F_id").Field("ID").Type(1, `,autoincrement`),
		)

		col := columns.AutoIncrement()
		require.NotNil(t, col)
		require.Equal(t, "f_id", col.Name)
	}

	require.Nil(t, columns.F("ID2"))
	require.Equal(t, 0, MustCols(columns.Fields("ID2")).Len())
	require.Equal(t, 1, MustCols(columns.Fields()).Len())
	require.Len(t, MustCols(columns.Fields("ID2")).List(), 0)
	require.Equal(t, 1, MustCols(columns.Cols("F_id")).Len())
	require.Equal(t, 1, MustCols(columns.Cols()).Len())
	require.Len(t, MustCols(columns.Cols("F_id")).List(), 1)
	require.Equal(t, []string{"ID"}, MustCols(columns.Cols("F_id")).FieldNames())
}

func MustCols(cols *Columns, err error) *Columns {
	return cols
}
