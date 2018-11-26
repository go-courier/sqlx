package builder

import (
	"testing"
)

func TestSelect(t *testing.T) {
	table := T("t")

	cases := map[string]struct {
		expr   SqlExpr
		expect SqlExpr
	}{
		"Select with modifier": {
			Select(nil, "DISTINCT").
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
				),
			Expr(
				"SELECT DISTINCT * FROM t WHERE f_a = ?",
				1,
			),
		},
		"SELECT simple": {
			Select(nil).
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
					Comment("comment"),
				),
			Expr(
				"SELECT * FROM t WHERE f_a = ? /* comment */",
				1,
			),
		},
		"SELECT with select expr": {
			Select(Col("F_a")).
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
				),
			Expr(
				"SELECT f_a FROM t WHERE f_a = ?",
				1,
			),
		},
		"Select for update": {
			Select(nil).From(
				table,
				Where(Col("F_a").Eq(1)),
				ForUpdate(),
			),
			Expr(
				"SELECT * FROM t WHERE f_a = ? FOR UPDATE",
				1,
			),
		},
		"Select with group by": {
			Select(nil).
				From(
					table,
					Where(Col("F_a").Eq(1)),
					GroupBy(Col("F_a")).
						Having(Col("F_a").Eq(1)),
				),
			Expr(
				"SELECT * FROM t WHERE f_a = ? GROUP BY f_a HAVING f_a = ?",
				1, 1,
			),
		},
		"Select with desc group by": {
			Select(nil).
				From(
					table,
					Where(Col("F_a").Eq(1)),
					GroupBy(DescOrder(Col("F_b"))),
				),
			Expr(
				"SELECT * FROM t WHERE f_a = ? GROUP BY (f_b) DESC",
				1,
			),
		},
		"Select with combined ordered group by ": {
			Select(nil).
				From(
					table,
					Where(Col("F_a").Eq(1)),
					GroupBy(AscOrder(Col("F_a")), DescOrder(Col("F_b"))),
				),
			Expr(
				"SELECT * FROM t WHERE f_a = ? GROUP BY (f_a) ASC,(f_b) DESC",
				1,
			),
		},
		"Select with limit": {
			Select(nil).
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
					Limit(1),
				),
			Expr(
				"SELECT * FROM t WHERE f_a = ? LIMIT 1",
				1,
			),
		},
		"Select with limit -1": {
			Select(nil).
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
					Limit(-1),
				),
			Expr(
				"SELECT * FROM t WHERE f_a = ?",
				1,
			),
		},
		"Select with limit with offset": {
			Select(nil).
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
					Limit(1).Offset(200),
				),
			Expr(
				"SELECT * FROM t WHERE f_a = ? LIMIT 1 OFFSET 200",
				1,
			),
		},
		"Select with order": {
			Select(nil).
				From(
					table,
					OrderBy(
						AscOrder(Col("F_a")),
						DescOrder(Col("F_b")),
					),
					Where(Col("F_a").Eq(1)),
				),
			Expr(
				"SELECT * FROM t WHERE f_a = ? ORDER BY (f_a) ASC,(f_b) DESC",
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
