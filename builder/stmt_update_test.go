package builder

import (
	"testing"
)

func TestStmtUpdate(t *testing.T) {
	table := T("t")

	cases := map[string]struct {
		expr   SqlExpr
		expect SqlExpr
	}{
		"Update simple": {
			Update(table).
				Set(
					Col("F_a").ValueBy(1),
					Col("F_b").ValueBy(2),
				).
				Where(
					Col("F_a").Eq(1),
					Comment("Comment"),
				),
			Expr(
				"UPDATE t SET F_a = ?, F_b = ? WHERE F_a = ? /* Comment */",
				1, 2, 1,
			),
		},
		"Update with limit": {
			Update(table).Set(
				Col("F_a").ValueBy(1),
			).Where(
				Col("F_a").Eq(1),
				Limit(1),
			),
			Expr(
				"UPDATE t SET F_a = ? WHERE F_a = ? LIMIT 1",
				1, 1,
			),
		},
		"Update with order": {
			Update(table).Set(
				Col("F_a").ValueBy(Col("F_a").Incr(1)),
				Col("F_b").ValueBy(Col("F_b").Desc(2)),
			).
				Where(
					Col("F_a").Eq(3),
					OrderBy(DescOrder(Col("F_b")), AscOrder(Col("F_a"))),
				),
			Expr(
				"UPDATE t SET F_a = F_a + ?, F_b = F_b - ? WHERE F_a = ? ORDER BY (F_b) DESC,(F_a) ASC",
				1, 2, 3,
			),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			queryArgsEqual(t, c.expect, c.expr)
		})
	}
}
