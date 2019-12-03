package builder

import (
	"context"
)

type OrderByAddition struct {
}

func (OrderByAddition) weight() additionWeight {
	return orderByStmt
}

func OrderBy(orders ...*order) *orderBy {
	return &orderBy{
		orders: orders,
	}
}

var _ Addition = (*orderBy)(nil)

type orderBy struct {
	OrderByAddition
	orders []*order
}

func (o *orderBy) IsNil() bool {
	return o == nil || len(o.orders) == 0
}

func (o *orderBy) Ex(ctx context.Context) *Ex {
	e := Expr("ORDER BY ")
	for i := range o.orders {
		if i > 0 {
			e.WriteRune(',')
		}
		e.WriteExpr(o.orders[i])
	}
	return e.Ex(ctx)
}

func AscOrder(target SqlExpr) *order {
	return &order{target: target, typ: "ASC"}
}

func DescOrder(target SqlExpr) *order {
	return &order{target: target, typ: "DESC"}
}

var _ SqlExpr = (*order)(nil)

type order struct {
	target SqlExpr
	typ    string
}

func (o *order) IsNil() bool {
	return o == nil || IsNilExpr(o.target)
}

func (o *order) Ex(ctx context.Context) *Ex {
	e := Expr("")

	e.WriteGroup(func(e *Ex) {
		e.WriteExpr(o.target)
	})

	if o.typ != "" {
		e.WriteRune(' ')
		e.WriteString(o.typ)
	}

	return e.Ex(ctx)
}
