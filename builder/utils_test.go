package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlattenArgs(t *testing.T) {
	tt := require.New(t)

	{
		q, args := FlattenArgs(`#ID IN (?)`, []int{28, 29, 30})
		tt.Equal("#ID IN (?,?,?)", q)
		tt.Equal(args, []interface{}{28, 29, 30})
	}
	{
		q, args := FlattenArgs(`#ID = (?)`, []byte(""))
		tt.Equal("#ID = (?)", q)
		tt.Equal(args, []interface{}{[]byte("")})
	}

	{
		q, args := FlattenArgs(`#ID = ?`, Expr("#ID + ?", 1))
		tt.Equal("#ID = #ID + ?", q)
		tt.Equal(args, []interface{}{1})
	}
}
