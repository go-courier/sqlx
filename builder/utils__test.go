package builder

import (
	"context"
	"strings"
	"testing"

	"github.com/onsi/gomega"
)

func TestValueMap(t *testing.T) {
	type User struct {
		ID       uint64 `db:"f_id"`
		Name     string `db:"f_name"`
		Username string `db:"f_username"`
	}

	user := User{
		ID: 123123213,
	}

	t.Run("#FieldValuesFromStructBy", func(t *testing.T) {
		gomega.NewWithT(t).Expect(FieldValuesFromStructBy(user, []string{})).To(gomega.HaveLen(0))

		values := FieldValuesFromStructBy(user, []string{"ID"})

		gomega.NewWithT(t).Expect(values).To(gomega.Equal(FieldValues{
			"ID": user.ID,
		}))
	})

	t.Run("#FieldValuesFromStructBy", func(t *testing.T) {
		gomega.NewWithT(t).Expect(FieldValuesFromStructByNonZero(user)).
			To(gomega.Equal(FieldValues{
				"ID": user.ID,
			}))

		gomega.NewWithT(t).Expect(FieldValuesFromStructByNonZero(user, "Username")).
			To(gomega.Equal(FieldValues{
				"ID":       user.ID,
				"Username": user.Username,
			}))
	})

	t.Run("#GetColumnName", func(t *testing.T) {
		gomega.NewWithT(t).Expect(GetColumnName("Text", "")).To(gomega.Equal("f_text"))
		gomega.NewWithT(t).Expect(GetColumnName("Text", ",size=256")).To(gomega.Equal("f_text"))
		gomega.NewWithT(t).Expect(GetColumnName("Text", "f_text2")).To(gomega.Equal("f_text2"))
		gomega.NewWithT(t).Expect(GetColumnName("Text", "f_text2,default=''")).To(gomega.Equal("f_text2"))
	})
}

func TestParseDef(t *testing.T) {
	t.Run("index with Field Names", func(t *testing.T) {

		i := ParseIndexDefine("index i_xxx/BTREE Name")

		gomega.NewWithT(t).Expect(i).To(gomega.Equal(&IndexDefine{
			Kind:   "index",
			Name:   "i_xxx",
			Method: "BTREE",
			IndexDef: IndexDef{
				FieldNames: []string{"Name"},
			},
		}))
	})

	t.Run("primary with Field Names", func(t *testing.T) {

		i := ParseIndexDefine("primary ID Name")

		gomega.NewWithT(t).Expect(i).To(gomega.Equal(&IndexDefine{
			Kind: "primary",
			IndexDef: IndexDef{
				FieldNames: []string{"ID", "Name"},
			},
		}))
	})

	t.Run("index with expr", func(t *testing.T) {
		i := ParseIndexDefine("index i_xxx USING GIST (#TEST gist_trgm_ops)")

		gomega.NewWithT(t).Expect(i).To(gomega.Equal(&IndexDefine{
			Kind: "index",
			Name: "i_xxx",
			IndexDef: IndexDef{
				Expr: "USING GIST (#TEST gist_trgm_ops)",
			},
		}))
	})
}

type User struct {
	ID       uint64 `db:"f_id"`
	Name     string `db:"f_name"`
	Username string `db:"f_username"`
}

func (User) TableName() string {
	return "t_user"
}

type OrgUser struct {
	OrgID  uint64 `db:"f_org_id"`
	UserID uint64 `db:"f_user_id"`
}

func (OrgUser) TableName() string {
	return "t_org_user"
}

type Org struct {
	ID   uint64 `db:"f_id"`
	Name string `db:"f_name"`
}

func (Org) TableName() string {
	return "t_org"
}

type OrgUserAll struct {
	OrgUser
	User User `json:"user"`
	Org  Org  `json:"org"`
}

func TestColumnsByStruct(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		q := ColumnsByStruct(&User{}).Ex(context.Background()).Query()
		gomega.NewWithT(t).Expect(q).To(gomega.Equal("t_user.f_id AS t_user__f_id, t_user.f_name AS t_user__f_name, t_user.f_username AS t_user__f_username"))
	})

	t.Run("joined", func(t *testing.T) {
		q := ColumnsByStruct(&OrgUserAll{}).Ex(context.Background()).Query()

		for _, g := range strings.Split(q, ", ") {
			t.Log(g)
		}

		gomega.NewWithT(t).Expect(q).To(gomega.Equal("t_org_user.f_org_id AS t_org_user__f_org_id, t_org_user.f_user_id AS t_org_user__f_user_id, t_user.f_id AS t_user__f_id, t_user.f_name AS t_user__f_name, t_user.f_username AS t_user__f_username, t_org.f_id AS t_org__f_id, t_org.f_name AS t_org__f_name"))
	})
}
