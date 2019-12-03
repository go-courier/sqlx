package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
)

func TestGroupBy(t *testing.T) {
	table := T("T")

	t.Run("select group by", testingx.It(func(t *testingx.T) {
		t.Expect(
			Select(nil).
				From(
					table,
					Where(Col("F_a").Eq(1)),
					GroupBy(Col("F_a")).
						Having(Col("F_a").Eq(1)),
				),
		).To(BeExpr(
			`
SELECT * FROM T
WHERE f_a = ?
GROUP BY f_a HAVING f_a = ?
`,
			1, 1,
		))
	}))

	t.Run("select desc group by", testingx.It(func(t *testingx.T) {
		t.Expect(
			Select(nil).
				From(
					table,
					Where(Col("F_a").Eq(1)),
					GroupBy(AscOrder(Col("F_a")), DescOrder(Col("F_b"))),
				),
		).To(BeExpr(`
SELECT * FROM T
WHERE f_a = ?
GROUP BY (f_a) ASC,(f_b) DESC
`,
			1,
		))
	}))

	t.Run("select multi group by", testingx.It(func(t *testingx.T) {
		t.Expect(
			Select(nil).
				From(
					table,
					Where(Col("F_a").Eq(1)),
					GroupBy(AscOrder(Col("F_a")), DescOrder(Col("F_b"))),
				),
		).To(BeExpr(
			`
SELECT * FROM T
WHERE f_a = ?
GROUP BY (f_a) ASC,(f_b) DESC
`,
			1,
		))
	}))
}
