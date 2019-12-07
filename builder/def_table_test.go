package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
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

	t.Run("replace table", testingx.It(func(t *testingx.T) {
		t.Expect(tUser.Expr("#.*")).To(buidertestingutils.BeExpr("t_user.*"))
	}))

	t.Run("replace table col by field", testingx.It(func(t *testingx.T) {
		t.Expect(tUser.Expr("#ID = #ID + 1")).To(buidertestingutils.BeExpr("f_id = f_id + 1"))
	}))

	t.Run("replace table col by field for function", testingx.It(func(t *testingx.T) {
		t.Expect(tUser.Expr("COUNT(#ID)")).To(buidertestingutils.BeExpr("COUNT(f_id)"))
	}))

	t.Run("could handle context", testingx.It(func(t *testingx.T) {
		t.Expect(
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
	}))
}
