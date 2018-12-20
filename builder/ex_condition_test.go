package builder

import (
	"testing"
)

func TestBuilderCond(t *testing.T) {
	cases := map[string]struct {
		expr   SqlExpr
		expect SqlExpr
	}{
		"CondRules with nil": {
			NewCondRules(),
			nil,
		},
		"CondRules with all false": {
			NewCondRules().
				When(false, Col("a").Eq(1)).
				When(false, Col("b").Like(`g`)).
				When(false, Col("b").Like(`g`)),
			nil,
		},
		"CondRules": {
			NewCondRules().
				When(true, Col("a").Eq(1)).
				When(true, Col("b").Like(`g`)).
				When(false, Col("b").Like(`g`)),
			Expr(
				"(a = ?) AND (b LIKE ?)",
				1, "%g%",
			),
		},
		"Chain Condition": {
			Col("a").Eq(1).
				And(Col("b").LeftLike("c")).
				Or(Col("a").Eq(2)).
				Xor(Col("b").RightLike("g")),
			Expr(
				"(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
				1, "%c", 2, "g%",
			),
		},
		"Compose Condition": {
			Xor(
				Or(
					And(
						Col("a").Eq(1),
						Col("b").Like("c"),
					),
					Col("a").Eq(2),
				),
				Col("b").Like("g"),
			),
			Expr(
				"(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
				1, "%c%", 2, "%g%",
			),
		},
		"Skip nil": {
			Xor(
				Col("a").In(),
				Or(
					Col("a").NotIn(),
					And(
						nil,
						Col("a").Eq(1),
						Col("b").Like("c"),
					),
					Col("a").Eq(2),
				),
				Col("b").Like("g"),
			),
			Expr(
				"(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
				1, "%c%", 2, "%g%",
			),
		},
		"XOR and Or": {
			Xor(
				Col("a").In(),
				Or(
					Col("a").NotIn(),
					And(
						nil,
						Col("a").Eq(1),
						Col("b").Like("c"),
					),
					Col("a").Eq(2),
				),
				Col("b").Like("g"),
			),
			Expr(
				"(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
				1, "%c%", 2, "%g%",
			),
		},
		"XOR": {

			Xor(
				Col("a").Eq(1),
				Col("b").Like("g"),
			),
			Expr(
				"(a = ?) XOR (b LIKE ?)",
				1, "%g%",
			),
		},
		"Like": {
			Col("d").Like("e"),
			Expr(
				"d LIKE ?",
				"%e%",
			),
		},
		"Not like": {
			Col("d").NotLike("e"),
			Expr(
				"d NOT LIKE ?",
				"%e%",
			),
		},
		"Equal": {
			Col("d").Eq("e"),
			Expr(
				"d = ?",
				"e",
			),
		},
		"Not Equal": {
			Col("d").Neq("e"),
			Expr(
				"d <> ?",
				"e",
			),
		},
		"In": {
			Col("d").In("e", "f"),
			Expr(
				"d IN (?,?)",
				"e", "f",
			),
		},
		"NotIn": {
			Col("d").NotIn("e", "f"),
			Expr(
				"d NOT IN (?,?)",
				"e", "f",
			),
		},
		"Less than": {
			Col("d").Lt(3),
			Expr(
				"d < ?",
				3,
			),
		},
		"Less or equal than": {
			Col("d").Lte(3),
			Expr(
				"d <= ?",
				3,
			),
		},
		"Greater than": {
			Col("d").Gt(3),
			Expr(
				"d > ?",
				3,
			),
		},
		"Greater or equal than": {
			Col("d").Gte(3),
			Expr(
				"d >= ?",
				3,
			),
		},
		"Between": {
			Col("d").Between(0, 2),
			Expr(
				"d BETWEEN ? AND ?",
				0, 2,
			),
		},
		"Not between": {
			Col("d").NotBetween(0, 2),
			Expr(
				"d NOT BETWEEN ? AND ?",
				0, 2,
			),
		},
		"Is null": {
			Col("d").IsNull(),
			Expr(
				"d IS NULL",
			),
		},
		"Is not null": {
			Col("d").IsNotNull(),
			Expr(
				"d IS NOT NULL",
			),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			queryArgsEqual(t, c.expect, c.expr)
		})
	}
}
