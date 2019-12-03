package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
)

func TestSelect(t *testing.T) {
	table := T("T")

	t.Run("select with modifier", testingx.It(func(t *testingx.T) {
		t.Expect(
			Select(nil, "DISTINCT").
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
				),
		).To(BeExpr(`
SELECT DISTINCT * FROM T
WHERE f_a = ?`, 1))
	}))

	t.Run("select simple", testingx.It(func(t *testingx.T) {
		t.Expect(
			Select(nil).
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
					Comment("comment"),
				),
		).To(BeExpr(`
SELECT * FROM T
WHERE f_a = ?
/* comment */
`, 1))
	}))

	t.Run("select with target", testingx.It(func(t *testingx.T) {
		t.Expect(
			Select(Col("F_a")).
				From(table,
					Where(
						Col("F_a").Eq(1),
					),
				),
		).To(BeExpr(`
SELECT f_a FROM T
WHERE f_a = ?`, 1))
	}))

	t.Run("select for update", testingx.It(func(t *testingx.T) {
		t.Expect(
			Select(nil).From(
				table,
				Where(Col("F_a").Eq(1)),
				ForUpdate(),
			),
		).To(BeExpr(
			`
SELECT * FROM T
WHERE f_a = ?
FOR UPDATE
`,
			1,
		))
	}))
}
