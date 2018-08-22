package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilderCond(t *testing.T) {
	table := T(DB("db"), "t")

	cases := []struct {
		name   string
		expr   SqlExpr
		result SqlExpr
	}{
		{
			"CondRules",
			NewCondRules().
				When(true, Col(table, "a").Eq(1)).
				When(true, Col(table, "b").Like(`g`)).
				When(false, Col(table, "b").Like(`g`)).
				ToCond(),
			Expr(
				"(`a` = ?) AND (`b` LIKE ?)",
				1, "%g%",
			),
		},
		{
			"Chain Condition",
			Col(table, "a").Eq(1).
				And(Col(table, "b").LeftLike("c")).
				Or(Col(table, "a").Eq(2)).
				Xor(Col(table, "b").RightLike("g")).Expr(),
			Expr(
				"(((`a` = ?) AND (`b` LIKE ?)) OR (`a` = ?)) XOR (`b` LIKE ?)",
				1, "%c", 2, "g%",
			),
		},
		{
			"Compose Condition",
			Xor(
				Or(
					And(
						Col(table, "a").Eq(1),
						Col(table, "b").Like("c"),
					),
					Col(table, "a").Eq(2),
				),
				Col(table, "b").Like("g"),
			).Expr(),
			Expr(
				"(((`a` = ?) AND (`b` LIKE ?)) OR (`a` = ?)) XOR (`b` LIKE ?)",
				1, "%c%", 2, "%g%",
			),
		},
		{
			"Skip nil",
			Xor(
				Col(table, "a").In(),
				Or(
					Col(table, "a").NotIn(),
					And(
						nil,
						Col(table, "a").Eq(1),
						Col(table, "b").Like("c"),
					),
					Col(table, "a").Eq(2),
				),
				Col(table, "b").Like("g"),
			).Expr(),
			Expr(
				"(((`a` = ?) AND (`b` LIKE ?)) OR (`a` = ?)) XOR (`b` LIKE ?)",
				1, "%c%", 2, "%g%",
			),
		}, {
			"XOR",
			Xor(
				Col(table, "a").In(),
				Or(
					Col(table, "a").NotIn(),
					And(
						nil,
						Col(table, "a").Eq(1),
						Col(table, "b").Like("c"),
					),
					Col(table, "a").Eq(2),
				),
				Col(table, "b").Like("g"),
			).Expr(),
			Expr(
				"(((`a` = ?) AND (`b` LIKE ?)) OR (`a` = ?)) XOR (`b` LIKE ?)",
				1, "%c%", 2, "%g%",
			),
		}, {
			"XOR",
			Xor(
				Col(table, "a").Eq(1),
				Col(table, "b").Like("g"),
			).Expr(),
			Expr(
				"(`a` = ?) XOR (`b` LIKE ?)",
				1, "%g%",
			),
		}, {
			"Like",
			Col(table, "d").Like("e").Expr(),
			Expr(
				"`d` LIKE ?",
				"%e%",
			),
		}, {
			"Not like",
			Col(table, "d").NotLike("e").Expr(),
			Expr(
				"`d` NOT LIKE ?",
				"%e%",
			),
		}, {
			"Equal",
			Col(table, "d").Eq("e").Expr(),
			Expr(
				"`d` = ?",
				"e",
			),
		}, {
			"Not Equal",
			Col(table, "d").Neq("e").Expr(),
			Expr(
				"`d` <> ?",
				"e",
			),
		}, {
			"In",
			Col(table, "d").In("e", "f").Expr(),
			Expr(
				"`d` IN (?,?)",
				"e", "f",
			),
		}, {
			"NotIn",
			Col(table, "d").NotIn("e", "f").Expr(),
			Expr(
				"`d` NOT IN (?,?)",
				"e", "f",
			),
		}, {
			"Less than",
			Col(table, "d").Lt(3).Expr(),
			Expr(
				"`d` < ?",
				3,
			),
		}, {
			"Less or equal than",
			Col(table, "d").Lte(3).Expr(),
			Expr(
				"`d` <= ?",
				3,
			),
		}, {
			"Greater than",
			Col(table, "d").Gt(3).Expr(),
			Expr(
				"`d` > ?",
				3,
			),
		}, {
			"Greater or equal than",
			Col(table, "d").Gte(3).Expr(),
			Expr(
				"`d` >= ?",
				3,
			),
		}, {
			"Between",
			Col(table, "d").Between(0, 2).Expr(),
			Expr(
				"`d` BETWEEN ? AND ?",
				0, 2,
			),
		}, {
			"Not between",
			Col(table, "d").NotBetween(0, 2).Expr(),
			Expr(
				"`d` NOT BETWEEN ? AND ?",
				0, 2,
			),
		}, {
			"Is null",
			Col(table, "d").IsNull().Expr(),
			Expr(
				"`d` IS NULL",
			),
		}, {
			"Is not null",
			Col(table, "d").IsNotNull().Expr(),
			Expr(
				"`d` IS NOT NULL",
			),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, ExprFrom(c.expr), ExprFrom(c.result))
		})
	}
}
