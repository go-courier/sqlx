package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStmtUpdate(t *testing.T) {
	table := T(DB("db"), "t")

	cases := []struct {
		name   string
		expr   SqlExpr
		result SqlExpr
	}{
		{
			"Update with modifier",
			Update(table, IGNORE).
				Where(Col(table, "F_a").Eq(1)).
				Set(Col(table, "F_a").ValueBy(1)),
			Expr(
				"UPDATE IGNORE `db`.`t` SET `F_a` = ? WHERE `F_a` = ?",
				1, 1,
			),
		}, {
			"Update simple",
			Update(table).
				Set(
					Col(table, "F_a").ValueBy(1),
					Col(table, "F_b").ValueBy(2),
				).
				Where(
					Col(table, "F_a").Eq(1),
					Comment("Comment"),
				),
			Expr(
				"UPDATE `db`.`t` SET `F_a` = ?, `F_b` = ? WHERE `F_a` = ? /* Comment */",
				1, 2, 1,
			),
		}, {
			"Update with limit",
			Update(table).Set(
				Col(table, "F_a").ValueBy(1),
			).Where(
				Col(table, "F_a").Eq(1),
				Limit(1),
			),
			Expr(
				"UPDATE `db`.`t` SET `F_a` = ? WHERE `F_a` = ? LIMIT 1",
				1, 1,
			),
		}, {
			"Update with order",
			Update(table).Set(
				Col(table, "F_a").ValueBy(Col(table, "F_a").Incr(1)),
				Col(table, "F_b").ValueBy(Col(table, "F_b").Desc(2)),
			).
				Where(
					Col(table, "F_a").Eq(3),
					OrderBy(DescOrder(Col(table, "F_b")), AscOrder(Col(table, "F_a"))),
				),
			Expr(
				"UPDATE `db`.`t` SET `F_a` = `F_a` + ?, `F_b` = `F_b` - ? WHERE `F_a` = ? ORDER BY (`F_b`) DESC, (`F_a`) ASC",
				1, 2, 3,
			),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, ExprFrom(c.result), ExprFrom(c.expr))
		})
	}
}
