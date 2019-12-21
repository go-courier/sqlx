package builder

import (
	"context"
)

type WhereAddition struct{}

func (WhereAddition) AdditionType() AdditionType {
	return AdditionWhere
}

func Where(c SqlCondition) *where {
	return &where{
		condition: c,
	}
}

var _ Addition = (*where)(nil)

type where struct {
	WhereAddition
	condition SqlCondition
}

func (w *where) IsNil() bool {
	return w == nil || IsNilExpr(w.condition)
}

func (w *where) Ex(ctx context.Context) *Ex {
	e := Expr("WHERE ")
	e.WriteExpr(w.condition)
	return e.Ex(ctx)
}
