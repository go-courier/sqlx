package builder

import (
	"context"
	"strings"
)

type JoinAddition struct{}

func (JoinAddition) AdditionType() AdditionType {
	return AdditionJoin
}

func Join(table *Table, prefixes ...string) *join {
	return &join{
		prefix: strings.Join(prefixes, " "),
		target: table,
	}
}

func InnerJoin(table *Table) *join {
	return Join(table, "INNER")
}

func LeftJoin(table *Table) *join {
	return Join(table, "LEFT")
}

func RightJoin(table *Table) *join {
	return Join(table, "RIGHT")
}

func FullJoin(table *Table) *join {
	return Join(table, "FULL")
}

func CrossJoin(table *Table) *join {
	return Join(table, "CROSS")
}

var _ Addition = (*join)(nil)

type join struct {
	prefix         string
	target         *Table
	joinCondition  SqlCondition
	joinColumnList []*Column
	JoinAddition
}

func (j join) On(joinCondition SqlCondition) *join {
	j.joinCondition = joinCondition
	return &j
}

func (j join) Using(joinColumnList ...*Column) *join {
	j.joinColumnList = joinColumnList
	return &j
}

func (j *join) IsNil() bool {
	return j == nil || IsNilExpr(j.target) || (j.prefix != "CROSS" && IsNilExpr(j.joinCondition) && len(j.joinColumnList) == 0)
}

func (j *join) Ex(ctx context.Context) *Ex {
	t := "JOIN "
	if j.prefix != "" {
		t = j.prefix + " " + t
	}

	e := Expr(t)

	e.WriteExpr(j.target)

	if !(IsNilExpr(j.joinCondition)) {
		e.WriteExpr(ExprBy(func(ctx context.Context) *Ex {
			ex := Expr(" ON ")
			ex.WriteExpr(j.joinCondition)
			return ex.Ex(ctx)
		}))
	}

	if len(j.joinColumnList) > 0 {
		e.WriteExpr(ExprBy(func(ctx context.Context) *Ex {
			ex := Expr(" USING ")

			ex.WriteGroup(func(e *Ex) {
				for i := range j.joinColumnList {
					if i != 0 {
						ex.WriteString(", ")
					}
					ex.WriteExpr(j.joinColumnList[i])
				}
			})

			return ex.Ex(ContextWithToggles(ctx, Toggles{
				ToggleMultiTable: false,
			}))
		}))
	}

	return e.Ex(ctx)
}
