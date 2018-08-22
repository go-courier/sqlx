package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabase(t *testing.T) {
	tt := require.New(t)

	db := DB("db")

	tt.Nil(db.Table("t"))

	table := T(db, "t")
	db.Register(table)
	tt.NotNil(db.Table("t"))
	tt.Equal([]string{"t"}, db.TableNames())

	cases := []struct {
		name   string
		expr   SqlExpr
		result SqlExpr
	}{
		{
			"Drop Database",
			DropDatabase(db),
			Expr("DROP DATABASE `db`"),
		},
		{
			"Create Database if not exists",
			CreateDatebaseIfNotExists(db),
			Expr("CREATE DATABASE IF NOT EXISTS `db`"),
		},
		{
			"Create Database",
			CreateDatebase(db),
			Expr("CREATE DATABASE `db`"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, ExprFrom(c.expr), ExprFrom(c.result))
		})
	}
}
