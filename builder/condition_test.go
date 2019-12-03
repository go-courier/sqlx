package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
)

func TestConditions(t *testing.T) {
	t.Run("Chain Condition", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("a").Eq(1).
				And(nil).
				And(Col("b").LeftLike("text")).
				Or(Col("a").Eq(2)).
				Xor(Col("b").RightLike("g")),
		).To(BeExpr(
			"(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
			1, "%text", 2, "g%",
		))
	}))

	t.Run("Compose Condition", testingx.It(func(t *testingx.T) {
		t.Expect(
			Xor(
				Or(
					And(
						(*Condition)(nil),
						(*Condition)(nil),
						(*Condition)(nil),
						(*Condition)(nil),
						Col("c").In(1, 2),
						Col("c").In([]int{3, 4}),
						Col("a").Eq(1),
						Col("b").Like("text"),
					),
					Col("a").Eq(2),
				),
				Col("b").Like("g"),
			),
		).To(BeExpr(
			"(((c IN (?,?)) AND (c IN (?,?)) AND (a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
			1, 2, 3, 4, 1, "%text%", 2, "%g%",
		))
	}))

	t.Run("skip nil", testingx.It(func(t *testingx.T) {
		t.Expect(
			Xor(
				Col("a").In(),
				Or(
					Col("a").NotIn(),
					And(
						nil,
						Col("a").Eq(1),
						Col("b").Like("text"),
					),
					Col("a").Eq(2),
				),
				Col("b").Like("g"),
			),
		).To(BeExpr(
			"(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
			1, "%text%", 2, "%g%",
		))
	}))

	t.Run("XOR and OR", testingx.It(func(t *testingx.T) {
		t.Expect(
			Xor(
				Col("a").In(),
				Or(
					Col("a").NotIn(),
					And(
						nil,
						Col("a").Eq(1),
						Col("b").Like("text"),
					),
					Col("a").Eq(2),
				),
				Col("b").Like("g"),
			),
		).To(BeExpr(
			"(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
			1, "%text%", 2, "%g%",
		))
	}))

	t.Run("XOR", testingx.It(func(t *testingx.T) {
		t.Expect(
			Xor(
				Col("a").Eq(1),
				Col("b").Like("g"),
			),
		).To(BeExpr(
			"(a = ?) XOR (b LIKE ?)",
			1, "%g%",
		))
	}))

	t.Run("Like", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").Like("e"),
		).To(BeExpr(
			"d LIKE ?",
			"%e%",
		))
	}))

	t.Run("Not like", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").NotLike("e"),
		).To(BeExpr(
			"d NOT LIKE ?",
			"%e%",
		))
	}))

	t.Run("Equal", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").Eq("e"),
		).To(BeExpr(
			"d = ?",
			"e",
		))
	}))

	t.Run("Not Equal", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").Neq("e"),
		).To(BeExpr(
			"d <> ?",
			"e",
		))
	}))

	t.Run("In", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").In("e", "f"),
		).To(BeExpr(
			"d IN (?,?)",
			"e", "f",
		))
	}))

	t.Run("NotIn", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").NotIn("e", "f"),
		).To(BeExpr(
			"d NOT IN (?,?)",
			"e", "f",
		))
	}))

	t.Run("Less than", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").Lt(3),
		).To(BeExpr(
			"d < ?",
			3,
		))
	}))

	t.Run("Less or equal than", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").Lte(3),
		).To(BeExpr(
			"d <= ?",
			3,
		))
	}))

	t.Run("Greater than", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").Gt(3),
		).To(BeExpr(
			"d > ?",
			3,
		))
	}))

	t.Run("Greater or equal than", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").Gte(3),
		).To(BeExpr(
			"d >= ?",
			3,
		))
	}))

	t.Run("Between", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").Between(0, 2),
		).To(BeExpr(
			"d BETWEEN ? AND ?",
			0, 2,
		))
	}))

	t.Run("Not between", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").NotBetween(0, 2),
		).To(BeExpr(
			"d NOT BETWEEN ? AND ?",
			0, 2,
		))
	}))

	t.Run("Is null", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").IsNull(),
		).To(BeExpr(
			"d IS NULL",
		))
	}))

	t.Run("Is not null", testingx.It(func(t *testingx.T) {
		t.Expect(
			Col("d").IsNotNull(),
		).To(BeExpr(
			"d IS NOT NULL",
		))
	}))
}
