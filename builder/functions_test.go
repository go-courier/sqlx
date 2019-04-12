package builder

import (
	"testing"
)

func TestFunc(t *testing.T) {
	cases := map[string]struct {
		expr   SqlExpr
		expect SqlExpr
	}{
		"Nil": {
			Func(""),
			nil,
		},
		"COUNT": {
			Count(),
			Expr("COUNT(1)"),
		},
		"AVG": {
			Avg(),
			Expr("AVG(*)"),
		},
		"DISTINCT": {
			Distinct(),
			Expr("DISTINCT(*)"),
		},
		"MIN": {
			Min(),
			Expr("MIN(*)"),
		},
		"Max": {
			Max(),
			Expr("MAX(*)"),
		},
		"First": {
			First(),
			Expr("FIRST(*)"),
		},
		"Last": {
			Last(),
			Expr("LAST(*)"),
		},
		"Sum": {
			Sum(),
			Expr("SUM(*)"),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			queryArgsEqual(t, c.expect, c.expr)
		})
	}
}
