package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelect(t *testing.T) {
	table := T(DB("db"), "t")

	cases := []struct {
		name   string
		expr   SqlExpr
		result SqlExpr
	}{
		{
			"Select with modifier",
			Select(nil, DISTINCT).
				From(
					table,
					Where(
						Col(table, "F_a").Eq(1),
					),
				),
			Expr(
				"SELECT DISTINCT * FROM `db`.`t` WHERE `F_a` = ?",
				1,
			),
		}, {
			"SELECT simple",
			Select(nil).
				From(
					table,
					Where(
						Col(table, "F_a").Eq(1),
					),
					Comment("comment"),
				),
			Expr(
				"SELECT * FROM `db`.`t` WHERE `F_a` = ? /* comment */",
				1,
			),
		}, {
			"SELECT with select expr",
			Select(Col(table, "F_a")).
				From(
					table,
					Where(
						Col(table, "F_a").Eq(1),
					),
				),
			Expr(
				"SELECT `F_a` FROM `db`.`t` WHERE `F_a` = ?",
				1,
			),
		}, {
			"Select for update",
			Select(nil).From(
				table,
				Where(Col(table, "F_a").Eq(1)),
				ForUpdate(),
			),
			Expr(
				"SELECT * FROM `db`.`t` WHERE `F_a` = ? FOR UPDATE",
				1,
			),
		}, {
			"Select with group by",
			Select(nil).
				From(
					table,
					Where(Col(table, "F_a").Eq(1)),
					GroupBy(Col(table, "F_a")).
						Having(Col(table, "F_a").Eq(1)),
				),
			Expr(
				"SELECT * FROM `db`.`t` WHERE `F_a` = ? GROUP BY `F_a` HAVING `F_a` = ?",
				1, 1,
			),
		}, {
			"Select with group by with rollup",
			Select(nil).
				From(
					table,
					Where(Col(table, "F_a").Eq(1)),
					GroupBy(Col(table, "F_a")).
						WithRollup().
						Having(Col(table, "F_a").Eq(1)),
				),
			Expr(
				"SELECT * FROM `db`.`t` WHERE `F_a` = ? GROUP BY `F_a` WITH ROLLUP HAVING `F_a` = ?",
				1, 1,
			),
		}, {
			"Select with desc group by",
			Select(nil).
				From(
					table,
					Where(Col(table, "F_a").Eq(1)),
					GroupBy(DescOrder(Col(table, "F_b"))),
				),
			Expr(
				"SELECT * FROM `db`.`t` WHERE `F_a` = ? GROUP BY (`F_b`) DESC",
				1,
			),
		}, {
			"Select with combined ordered group by ",
			Select(nil).
				From(
					table,
					Where(Col(table, "F_a").Eq(1)),
					GroupBy(AscOrder(Col(table, "F_a")), DescOrder(Col(table, "F_b"))),
				),
			Expr(
				"SELECT * FROM `db`.`t` WHERE `F_a` = ? GROUP BY (`F_a`) ASC, (`F_b`) DESC",
				1,
			),
		}, {
			"Select with limit",
			Select(nil).
				From(
					table,
					Where(
						Col(table, "F_a").Eq(1),
					),
					Limit(1),
				),
			Expr(
				"SELECT * FROM `db`.`t` WHERE `F_a` = ? LIMIT 1",
				1,
			),
		}, {
			"Select with limit -1",
			Select(nil).
				From(
					table,
					Where(
						Col(table, "F_a").Eq(1),
					),
					Limit(-1),
				),
			Expr(
				"SELECT * FROM `db`.`t` WHERE `F_a` = ?",
				1,
			),
		}, {
			"Select with limit with offset",
			Select(nil).
				From(
					table,
					Where(
						Col(table, "F_a").Eq(1),
					),
					Limit(1).Offset(200),
				),
			Expr(
				"SELECT * FROM `db`.`t` WHERE `F_a` = ? LIMIT 1 OFFSET 200",
				1,
			),
		}, {
			"Select with order",
			Select(nil).
				From(
					table,
					OrderBy(AscOrder(Col(table, "F_a")), DescOrder(Col(table, "F_b"))),
					Where(Col(table, "F_a").Eq(1)),
				),
			Expr(
				"SELECT * FROM `db`.`t` WHERE `F_a` = ? ORDER BY (`F_a`) ASC, (`F_b`) DESC",
				1,
			),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, ExprFrom(c.result), ExprFrom(c.expr))
		})
	}
}
