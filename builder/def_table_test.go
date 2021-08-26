package builder_test

import (
	"context"
	"testing"

	"github.com/go-courier/sqlx/v2/connectors/postgresql"

	. "github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/onsi/gomega"
)

func TestTable_Expr(t *testing.T) {
	tUser := T("t_user",
		Col("f_id").Field("ID").Type(uint64(0), ",autoincrement"),
		Col("f_name").Field("Name").Type("", ",size=128,default=''"),
	)

	tUserRole := T("t_user_role",
		Col("f_id").Field("ID").Type(uint64(0), ",autoincrement"),
		Col("f_user_id").Field("UserID").Type(uint64(0), ""),
	)

	t.Run("replace table", func(t *testing.T) {
		gomega.NewWithT(t).Expect(tUser.Expr("#.*")).To(buidertestingutils.BeExpr("t_user.*"))
	})
	t.Run("replace table col by field", func(t *testing.T) {
		gomega.NewWithT(t).Expect(tUser.Expr("#ID = #ID + 1")).To(buidertestingutils.BeExpr("f_id = f_id + 1"))
	})
	t.Run("replace table col by field for function", func(t *testing.T) {
		gomega.NewWithT(t).Expect(tUser.Expr("COUNT(#ID)")).To(buidertestingutils.BeExpr("COUNT(f_id)"))
	})
	t.Run("could handle context", func(t *testing.T) {
		gomega.NewWithT(t).Expect(
			Select(nil).
				From(
					tUser,
					Where(
						AsCond(tUser.Expr("#ID > 1")),
					),
					Join(tUserRole).On(AsCond(tUser.Expr("#ID = ?", tUserRole.Expr("#UserID")))),
				),
		).To(buidertestingutils.BeExpr(`
SELECT * FROM t_user
JOIN t_user_role ON t_user.f_id = t_user_role.f_user_id
WHERE t_user.f_id > 1
`))
	})

	t.Run("diff", func(t *testing.T) {
		tUser := T("t_user",
			Col("f_id").Field("ID").Type(uint64(0), ",autoincrement"),
			Col("f_name").Field("Name").Type("", ",size=128,default=''"),
		)

		tUser2 := T("t_user",
			Col("f_id").Field("ID").Type(uint64(0), ",autoincrement"),
			Col("f_name").Field("Name").Type("", ",size=128,default=''"),
			Col("f_nickname").Field("Nickname").Type("", ",size=128,default=''"),
			PrimaryKey(Cols("f_id")),
		)

		exprList := tUser2.Diff(tUser, &postgresql.PostgreSQLConnector{})

		for _, expr := range exprList {
			t.Log(expr.Ex(context.Background()).Query())
		}
	})
}
