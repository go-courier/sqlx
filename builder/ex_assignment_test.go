package builder

import (
	"testing"
)

func TestAssignment(t *testing.T) {
	cases := map[string]struct {
		expr   SqlExpr
		expect SqlExpr
	}{
		"ColumnsAndValues": {
			ColumnsAndValues(Cols("a", "b"), 1, 2, 3, 4),
			Expr("(a,b) VALUES (?,?),(?,?)", 1, 2, 3, 4),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			queryArgsEqual(t, c.expect, c.expr)
		})
	}
}
