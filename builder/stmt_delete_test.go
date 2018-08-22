package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStmtDelete(t *testing.T) {
	table := T(DB("db"), "t")

	cases := []struct {
		name   string
		expr   SqlExpr
		result SqlExpr
	}{
		{
			"Delete with modifier",
			Delete(IGNORE).From(
				table,
				Where(Col(table, "F_a").Eq(1)),
			),
			Expr(
				"DELETE IGNORE FROM `db`.`t` WHERE `F_a` = ?",
				1,
			),
		}, {
			"Delete simple",
			Delete().From(table,
				Where(Col(table, "F_a").Eq(1)),
				Comment("Comment"),
			),
			Expr(
				"DELETE FROM `db`.`t` WHERE `F_a` = ? /* Comment */",
				1,
			),
		}, {
			"Delete with limit",
			Delete().From(table,
				Where(Col(table, "F_a").Eq(1)),
				Limit(1),
			),
			Expr(
				"DELETE FROM `db`.`t` WHERE `F_a` = ? LIMIT 1",
				1,
			),
		}, {
			"Delete with order",
			Delete().From(table,
				Where(Col(table, "F_a").Eq(1)),
				OrderBy(AscOrder(Col(table, "F_a")), DescOrder(Col(table, "F_b"))),
			),
			Expr(
				"DELETE FROM `db`.`t` WHERE `F_a` = ? ORDER BY (`F_a`) ASC, (`F_b`) DESC",
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
