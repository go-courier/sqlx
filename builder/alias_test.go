package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/onsi/gomega"
)

func TestAlias(t *testing.T) {
	t.Run("alias", func(t *testing.T) {
		gomega.NewWithT(t).Expect(Alias(Expr("f_id"), "id")).To(BeExpr("(f_id) AS id"))
	})
}
