package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValueMap_FromStructBy(t *testing.T) {
	tt := require.New(t)

	type User struct {
		ID       uint64 `db:"F_id" sql:"bigint(64) unsigned NOT NULL AUTO_INCREMENT"`
		Name     string `db:"F_name" sql:"varchar(255) NOT NULL"`
		Username string `db:"F_username" sql:"varchar(255) NOT NULL"`
	}

	user := User{
		ID: 123123213,
	}

	{
		fieldValues := FieldValuesFromStructBy(user, []string{})
		tt.Len(fieldValues, 0)
	}
	{
		fieldValues := FieldValuesFromStructBy(user, []string{"ID"})
		tt.Equal(fieldValues, FieldValues{
			"ID": user.ID,
		})
	}
}

func TestValueMap_FromStructWithEmptyFields(t *testing.T) {
	tt := require.New(t)

	type User struct {
		ID       uint64 `db:"F_id" sql:"bigint(64) unsigned NOT NULL AUTO_INCREMENT"`
		Name     string `db:"F_name" sql:"varchar(255) NOT NULL"`
		Username string `db:"F_username" sql:"varchar(255) NOT NULL"`
	}

	user := User{
		ID: 123123213,
	}

	{
		fieldValues := FieldValuesFromStructByNonZero(user)
		tt.Equal(fieldValues, FieldValues{
			"ID": user.ID,
		})
	}
	{
		fieldValues := FieldValuesFromStructByNonZero(user, "Username")
		tt.Equal(fieldValues, FieldValues{
			"ID":       user.ID,
			"Username": user.Username,
		})
	}
}
