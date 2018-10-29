package builder

import (
	"testing"
)

func TestStmtInsert(t *testing.T) {
	table := T("t")

	cases := map[string]struct {
		expr   SqlExpr
		expect SqlExpr
	}{
		"Insert simple": {
			Insert().
				Into(table, Comment("Comment")).
				Values(Cols("F_a", "F_b"), 1, 2),
			Expr(
				"INSERT INTO t (f_a,f_b) VALUES (?,?) /* Comment */",
				1, 2,
			),
		},
		"Insert with modifier": {
			Insert("IGNORE").
				Into(table).
				Values(Cols("F_a", "F_b"), 1, 2),
			Expr(
				"INSERT IGNORE INTO t (f_a,f_b) VALUES (?,?)",
				1, 2,
			),
		},
		"Insert multiple": {
			Insert().
				Into(table).
				Values(Cols("F_a", "F_b"), 1, 2, 1, 2, 1, 2),
			Expr(
				"INSERT INTO t (f_a,f_b) VALUES (?,?),(?,?),(?,?)",
				1, 2, 1, 2, 1, 2,
			),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			queryArgsEqual(t, c.expect, c.expr)
		})
	}
}
