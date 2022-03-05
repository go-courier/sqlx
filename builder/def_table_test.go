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
			Index("f_name", nil, "(#Name DESC NULLS LAST)"),
		)

		t.Run("from user to user2", func(t *testing.T) {
			exprList := tUser2.Diff(tUser, &postgresql.PostgreSQLConnector{})

			exprs := make([]string, len(exprList))
			for i, expr := range exprList {
				exprs[i] = expr.Ex(context.Background()).Query()
			}

			gomega.NewWithT(t).Expect(exprs).To(gomega.Equal([]string{
				"ALTER TABLE t_user ADD COLUMN f_nickname character varying(128) NOT NULL DEFAULT ''::character varying;",
				"ALTER TABLE t_user ADD PRIMARY KEY (f_id);",
				"CREATE INDEX t_user_f_name ON t_user (f_name DESC NULLS LAST);",
			}))
		})

		t.Run("from user2 to user1", func(t *testing.T) {
			exprList := tUser.Diff(tUser2, &postgresql.PostgreSQLConnector{})

			exprs := make([]string, len(exprList))
			for i, expr := range exprList {
				exprs[i] = expr.Ex(context.Background()).Query()
			}

			gomega.NewWithT(t).Expect(exprs).To(gomega.Equal([]string{
				"ALTER TABLE t_user DROP CONSTRAINT t_user_pkey;",
				"DROP INDEX IF EXISTS t_user_f_name;",
			}))
		})
	})
}
