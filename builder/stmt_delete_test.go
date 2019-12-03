package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
)

func TestStmtDelete(t *testing.T) {
	table := T("T")

	t.Run("delete", testingx.It(func(t *testingx.T) {
		t.Expect(
			Delete().From(table,
				Where(Col("F_a").Eq(1)),
				Comment("Comment"),
			),
		).To(BeExpr(`
DELETE FROM T
WHERE f_a = ?
/* Comment */
`, 1))
	}))
}
