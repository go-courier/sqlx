package builder

import (
	"testing"
)

func TestStmtDelete(t *testing.T) {
	table := T("t")

	cases := map[string]struct {
		expr   SqlExpr
		expect SqlExpr
	}{
		"Delete simple": {
			Delete().From(table,
				Where(Col("F_a").Eq(1)),
				Comment("Comment"),
			),
			Expr(
				"DELETE FROM t WHERE f_a = ? /* Comment */",
				1,
			),
		},
		"Delete with limit": {
			Delete().From(
				table,
				Where(Col("F_a").Eq(1)),
				Limit(1),
			),
			Expr(
				"DELETE FROM t WHERE f_a = ? LIMIT 1",
				1,
			),
		},
		"Delete with order": {
			Delete().From(
				table,
				Where(Col("F_a").Eq(1)),
				OrderBy(
					AscOrder(Col("F_a")),
					DescOrder(Col("F_b")),
				),
			),
			Expr(
				"DELETE FROM t WHERE f_a = ? ORDER BY (f_a) ASC,(f_b) DESC",
				1,
			),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			queryArgsEqual(t, c.expect, c.expr)
		})
	}
}
