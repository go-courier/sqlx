package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
)

func TestFunc(t *testing.T) {
	t.Run("invalid", testingx.It(func(t *testingx.T) {
		t.Expect(Func("")).To(BeExpr(""))
	}))

	t.Run("count", testingx.It(func(t *testingx.T) {
		t.Expect(Count()).To(BeExpr("COUNT(1)"))
	}))

	t.Run("AVG", testingx.It(func(t *testingx.T) {
		t.Expect(Avg()).To(BeExpr("AVG(*)"))
	}))
}
