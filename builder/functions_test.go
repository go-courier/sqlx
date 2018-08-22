package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFunc(t *testing.T) {
	cases := []struct {
		name   string
		expr   SqlExpr
		result SqlExpr
	}{
		{
			"Nil",
			Func(""),
			nil,
		},
		{
			"COUNT",
			Count(),
			Expr("COUNT(*)"),
		},
		{
			"AVG",
			Avg(),
			Expr("AVG(*)"),
		},
		{
			"DISTINCT",
			Distinct(),
			Expr("DISTINCT(*)"),
		},
		{
			"MIN",
			Min(),
			Expr("MIN(*)"),
		},
		{
			"Max",
			Max(),
			Expr("MAX(*)"),
		},
		{
			"First",
			First(),
			Expr("FIRST(*)"),
		},
		{
			"Last",
			Last(),
			Expr("LAST(*)"),
		},
		{
			"Sum",
			Sum(),
			Expr("SUM(*)"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, ExprFrom(c.result), ExprFrom(c.expr))
		})
	}
}
