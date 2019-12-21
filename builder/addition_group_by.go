package builder

import (
	"context"
)

type GroupByAddition struct {
}

func (GroupByAddition) AdditionType() AdditionType {
	return AdditionGroupBy
}

func GroupBy(groups ...SqlExpr) *groupBy {
	return &groupBy{
		groups: groups,
	}
}

var _ Addition = (*groupBy)(nil)

type groupBy struct {
	GroupByAddition

	groups []SqlExpr
	// HAVING
	havingCond SqlCondition
}

func (g groupBy) Having(cond SqlCondition) *groupBy {
	g.havingCond = cond
	return &g
}

func (g *groupBy) IsNil() bool {
	return g == nil || len(g.groups) == 0
}

func (g *groupBy) Ex(ctx context.Context) *Ex {
	expr := Expr("GROUP BY ")

	for i, group := range g.groups {
		if i > 0 {
			expr.WriteByte(',')
		}
		expr.WriteExpr(group)
	}

	if !(IsNilExpr(g.havingCond)) {
		expr.WriteString(" HAVING ")
		expr.WriteExpr(g.havingCond)
	}

	return expr.Ex(ctx)
}
