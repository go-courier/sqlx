package builder

import (
	"database/sql/driver"
	"fmt"
	"testing"
)

func TestEx(t *testing.T) {
	cases := map[string]struct {
		expr   SqlExpr
		expect SqlExpr
	}{
		"empty": {
			ExprFrom(nil),
			nil,
		},
		"expr": {
			ExprFrom(Expr("")),
			Expr(""),
		},
		"flatten for slice": {
			Expr(`#ID IN (?)`, []int{28, 29, 30}).Flatten(),
			Expr("#ID IN (?,?,?)", 28, 29, 30),
		},
		"flatten skip for bytes": {
			Expr(`#ID = (?)`, []byte("")).Flatten(),
			Expr("#ID = (?)", []byte("")),
		},
		"flatten SqlExpr": {
			Expr(`#ID = ?`, Expr("#ID + ?", 1)).Flatten(),
			Expr("#ID = #ID + ?", 1),
		},
		"flatten ValuerExpr": {
			Expr(`#Point = ?`, Point{1, 1}).Flatten(),
			Expr("#Point = ST_GeomFromText(?)", Point{1, 1}),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			queryArgsEqual(t, c.expect, c.expr)
		})
	}
}

type Point struct {
	X float64
	Y float64
}

func (Point) DataType(engine string) string {
	return "POINT"
}

func (Point) ValueEx() string {
	return `ST_GeomFromText(?)`
}

func (p Point) Value() (driver.Value, error) {
	return fmt.Sprintf("POINT(%v %v)", p.X, p.Y), nil
}
