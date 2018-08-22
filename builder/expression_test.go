package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpression(t *testing.T) {
	cases := []struct {
		name   string
		expr   SqlExpr
		result SqlExpr
	}{
		{
			"empty",
			ExprFrom(nil),
			nil,
		},
		{
			"expr",
			ExprFrom(Expr("")),
			Expr(""),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, ExprFrom(c.result), ExprFrom(c.expr))
		})
	}
}

func TestHolderRepeat(t *testing.T) {
	tt := require.New(t)
	tt.Equal("?,?,?,?,?", HolderRepeat(5))
}

func TestQuote(t *testing.T) {
	tt := require.New(t)
	tt.Equal("`name`", quote("name"))
}
