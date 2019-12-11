package builder

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestValueMap(t *testing.T) {
	type User struct {
		ID       uint64 `db:"F_id"`
		Name     string `db:"F_name"`
		Username string `db:"F_username"`
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
}
