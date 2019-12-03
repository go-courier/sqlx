package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
)

func TestAssignment(t *testing.T) {
	t.Run("ColumnsAndValues", testingx.It(func(t *testingx.T) {
		t.Expect(
			ColumnsAndValues(Cols("a", "b"), 1, 2, 3, 4),
		).To(BeExpr("(a,b) VALUES (?,?),(?,?)", 1, 2, 3, 4))
	}))
}
