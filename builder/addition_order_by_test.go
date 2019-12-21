package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/onsi/gomega"
)

func TestOrderBy(t *testing.T) {
	table := T("T")

	t.Run("select Order", func(t *testing.T) {
		gomega.NewWithT(t).Expect(
			Select(nil).
				From(
					table,
					OrderBy(
						AscOrder(Col("F_a")),
						DescOrder(Col("F_b")),
					),
					Where(Col("F_a").Eq(1)),
				),
		).To(BeExpr(
			`
SELECT * FROM T
WHERE f_a = ?
ORDER BY (f_a) ASC,(f_b) DESC
`,
			1,
		))
	})
}
