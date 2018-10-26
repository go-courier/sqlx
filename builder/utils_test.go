package builder

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func queryArgsEqual(t *testing.T, expect SqlExpr, actual SqlExpr) {
	e := ExprFrom(expect)
	a := ExprFrom(actual)

	if e == nil || a == nil {
		require.Equal(t, e, a)
	} else {
		e = e.Flatten()
		a = a.Flatten()

		require.Equal(t, e.Query(), a.Query())
		require.Equal(t, e.Args(), a.Args())
	}
}

func TestValueMap_FromStructBy(t *testing.T) {
	type User struct {
		ID       uint64 `db:"F_id"`
		Name     string `db:"F_name"`
		Username string `db:"F_username"`
	}

	user := User{
		ID: 123123213,
	}

	{
		fieldValues := FieldValuesFromStructBy(user, []string{})
		require.Len(t, fieldValues, 0)
	}
	{
		fieldValues := FieldValuesFromStructBy(user, []string{"ID"})
		require.Equal(t, fieldValues, FieldValues{
			"ID": user.ID,
		})
	}
}

func TestValueMap_FromStructWithEmptyFields(t *testing.T) {
	type User struct {
		ID       uint64 `db:"F_id"`
		Name     string `db:"F_name"`
		Username string `db:"F_username"`
	}

	user := User{
		ID: 123123213,
	}

	{
		fieldValues := FieldValuesFromStructByNonZero(user)
		require.Equal(t, fieldValues, FieldValues{
			"ID": user.ID,
		})
	}
	{
		fieldValues := FieldValuesFromStructByNonZero(user, "Username")
		require.Equal(t, fieldValues, FieldValues{
			"ID":       user.ID,
			"Username": user.Username,
		})
	}
}
