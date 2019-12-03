package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
)

func TestAlias(t *testing.T) {
	t.Run("alias", testingx.It(func(t *testingx.T) {
		t.Expect(Alias(Expr("f_id"), "id")).To(BeExpr("(f_id) AS id"))
	}))
}
