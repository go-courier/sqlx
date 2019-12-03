package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
)

func TestLimit(t *testing.T) {
	table := T("T")

	t.Run("select limit", testingx.It(func(t *testingx.T) {
		t.Expect(
			Select(nil).
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
					Limit(1),
				),
		).To(BeExpr(`
SELECT * FROM T
WHERE f_a = ?
LIMIT 1
`, 1))
	}))

	t.Run("select without limit", testingx.It(func(t *testingx.T) {
		t.Expect(
			Select(nil).
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
					Limit(-1),
				),
		).To(BeExpr(`
SELECT * FROM T
WHERE f_a = ?
`, 1,
		))
	}))

	t.Run("select limit and offset", testingx.It(func(t *testingx.T) {
		t.Expect(
			Select(nil).
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
					Limit(1).Offset(200),
				),
		).To(BeExpr(`
SELECT * FROM T
WHERE f_a = ?
LIMIT 1 OFFSET 200
`,
			1,
		))
	}))

}
