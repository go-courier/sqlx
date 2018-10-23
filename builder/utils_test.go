package builder

import (
	"database/sql/driver"
	"fmt"
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

	{
		q, args := FlattenArgs(`#Point = ?`, Point{1, 1})
		tt.Equal("#Point = ST_GeomFromText(?)", q)
		tt.Equal(args, []interface{}{Point{1, 1}})
	}
}

type Point struct {
	X float64
	Y float64
}

func (Point) ValueEx() string {
	return `ST_GeomFromText(?)`
}

func (p Point) Value() (driver.Value, error) {
	return fmt.Sprintf("POINT(%v %v)", p.X, p.Y), nil
}
