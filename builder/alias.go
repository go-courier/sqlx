package builder

import (
	"context"
)

func Alias(expr SqlExpr, name string) *exAlias {
	return &exAlias{
		name:    name,
		SqlExpr: expr,
	}
}

type exAlias struct {
	name string
	SqlExpr
}

func (alias *exAlias) IsNil() bool {
	return alias == nil || alias.name == "" || IsNilExpr(alias.SqlExpr)
}

func (alias *exAlias) Ex(ctx context.Context) *Ex {
	return Expr("(?) AS ?", alias.SqlExpr, Expr(alias.name)).Ex(ContextWithToggles(ctx, Toggles{
		ToggleNeedAutoAlias: false,
	}))
}

func MultiMayAutoAlias(columns ...SqlExpr) *exMayAutoAlias {
	return &exMayAutoAlias{
		columns: columns,
	}
}

type exMayAutoAlias struct {
	columns []SqlExpr
}

func (alias *exMayAutoAlias) IsNil() bool {
	return alias == nil || len(alias.columns) == 0
}

func (alias *exMayAutoAlias) Ex(ctx context.Context) *Ex {
	e := Expr("")

	RangeNotNilExpr(alias.columns, func(expr SqlExpr, i int) {
		if i > 0 {
			e.WriteString(", ")
		}
		e.WriteExpr(expr)
	})

	return e.Ex(ContextWithToggles(ctx, Toggles{
		ToggleNeedAutoAlias: true,
	}))
}
