package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/onsi/gomega"
)

func TestJoin(t *testing.T) {
	tUser := T("t_user",
		Col("f_id").Type(uint64(0), ",autoincrement"),
		Col("f_name").Type("", ",size=128,default=''"),
		Col("f_org_id").Type("", ",size=128,default=''"),
	)

	tOrg := T("t_org",
		Col("f_org_id").Type(uint64(0), ",autoincrement"),
		Col("f_org_name").Type("", ",size=128,default=''"),
	)

	t.Run("JOIN ON", func(t *testing.T) {
		gomega.NewWithT(t).Expect(
			Select(MultiWith(", ",
				Alias(tUser.Col("f_id"), "f_id"),
				Alias(tUser.Col("f_name"), "f_name"),
				Alias(tUser.Col("f_org_id"), "f_org_id"),
				Alias(tOrg.Col("f_org_name"), "f_org_name"),
			)).
				From(
					tUser,
					Join(Alias(tOrg, "t_org")).On(tUser.Col("f_org_id").Eq(tOrg.Col("f_org_id"))),
				)).
			To(BeExpr(
				`
SELECT t_user.f_id AS f_id, t_user.f_name AS f_name, t_user.f_org_id AS f_org_id, t_org.f_org_name AS f_org_name FROM t_user
JOIN t_org AS t_org ON t_user.f_org_id = t_org.f_org_id
`,
			))
	})
	t.Run("JOIN USING", func(t *testing.T) {
		gomega.NewWithT(t).Expect(
			Select(nil).
				From(
					tUser,
					Join(tOrg).Using(tUser.Col("f_org_id")),
				),
		).To(BeExpr(
			`
SELECT * FROM t_user
JOIN t_org USING (f_org_id)
`,
		))
	})
}
