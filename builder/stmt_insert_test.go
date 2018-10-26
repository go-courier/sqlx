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
				"INSERT INTO t (F_a,F_b) VALUES (?,?) /* Comment */",
				1, 2,
			),
		},
		"Insert with modifier": {
			Insert("IGNORE").
				Into(table).
				Values(Cols("F_a", "F_b"), 1, 2),
			Expr(
				"INSERT IGNORE INTO t (F_a,F_b) VALUES (?,?)",
				1, 2,
			),
		},
		"Insert multiple": {
			Insert().
				Into(table).
				Values(Cols("F_a", "F_b"), 1, 2, 1, 2, 1, 2),
			Expr(
				"INSERT INTO t (F_a,F_b) VALUES (?,?),(?,?),(?,?)",
				1, 2, 1, 2, 1, 2,
			),
		},
		//{
		//	"Insert on on duplicate key update",
		//	Insert().Into(
		//		table,
		//		OnDuplicateKeyUpdate(Col("F_b").ValueBy(2)),
		//	).
		//		Values(Cols("F_a", "F_b"), 1, 2),
		//	Expr(
		//		"INSERT INTO t (F_a,F_b) VALUES (?,?) ON DUPLICATE KEY UPDATE F_b = ?",
		//		1, 2, 2,
		//	),
		//},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			queryArgsEqual(t, c.expect, c.expr)
		})
	}
}
