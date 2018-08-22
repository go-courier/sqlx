package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStmtInsert(t *testing.T) {
	table := T(DB("db"), "t")

	cases := []struct {
		name   string
		expr   SqlExpr
		result SqlExpr
	}{
		{
			"Insert by set",
			Insert().
				Into(table).
				Set(Col(table, "F_a").ValueBy(1)),
			Expr(
				"INSERT INTO `db`.`t` SET `F_a` = ?",
				1,
			),
		},
		{
			"Insert simple",
			Insert().
				Into(table, Comment("Comment")).
				Values(Cols(table, "F_a", "F_b"), 1, 2),
			Expr(
				"INSERT INTO `db`.`t` (`F_a`,`F_b`) VALUES (?,?) /* Comment */",
				1, 2,
			),
		}, {
			"Insert with modifier",
			Insert(IGNORE).Into(table).
				Values(Cols(table, "F_a", "F_b"), 1, 2).
				Expr(),
			Expr(
				"INSERT IGNORE INTO `db`.`t` (`F_a`,`F_b`) VALUES (?,?)",
				1, 2,
			),
		}, {
			"Insert multiple",
			Insert().Into(table).
				Values(Cols(table, "F_a", "F_b"), 1, 2, 1, 2, 1, 2),
			Expr(
				"INSERT INTO `db`.`t` (`F_a`,`F_b`) VALUES (?,?),(?,?),(?,?)",
				1, 2, 1, 2, 1, 2,
			),
		}, {
			"Insert on on duplicate key update",
			Insert().Into(
				table,
				OnDuplicateKeyUpdate(Col(table, "F_b").ValueBy(2)),
			).
				Values(Cols(table, "F_a", "F_b"), 1, 2),
			Expr(
				"INSERT INTO `db`.`t` (`F_a`,`F_b`) VALUES (?,?) ON DUPLICATE KEY UPDATE `F_b` = ?",
				1, 2, 2,
			),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, ExprFrom(c.result), ExprFrom(c.expr))
		})
	}
}
