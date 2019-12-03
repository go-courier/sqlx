package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
)

func TestStmtInsert(t *testing.T) {
	table := T("T")

	t.Run("insert with modifier", testingx.It(func(t *testingx.T) {
		t.Expect(
			Insert("IGNORE").
				Into(table).
				Values(Cols("F_a", "F_b"), 1, 2),
		).To(BeExpr("INSERT IGNORE INTO T (f_a,f_b) VALUES (?,?)",
			1, 2))
	}))

	t.Run("insert simple", testingx.It(func(t *testingx.T) {
		t.Expect(
			Insert().
				Into(table, Comment("Comment")).
				Values(Cols("f_a", "F_b"), 1, 2),
		).To(BeExpr(`
INSERT INTO T (f_a,f_b) VALUES (?,?)
/* Comment */
`, 1, 2))
	}))

	t.Run("multiple insert", testingx.It(func(t *testingx.T) {
		t.Expect(
			Insert().
				Into(table).
				Values(Cols("F_a", "F_b"), 1, 2, 1, 2, 1, 2),
		).To(BeExpr("INSERT INTO T (f_a,f_b) VALUES (?,?),(?,?),(?,?)", 1, 2, 1, 2, 1, 2))
	}))
}
