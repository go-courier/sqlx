package builder

import (
	"sort"
)

type Additions []Addition

type Addition interface {
	SqlExpr
	AdditionType() AdditionType
}

type AdditionType int

const (
	AdditionJoin AdditionType = iota
	AdditionWhere
	AdditionGroupBy
	AdditionCombination
	AdditionOrderBy
	AdditionLimit
	AdditionOnConflict
	AdditionOther
	AdditionComment
)

func WriteAdditions(e *Ex, additions ...Addition) {
	finalAdditions := Additions{}
	for i := range additions {
		if IsNilExpr(additions[i]) {
			continue
		}
		finalAdditions = append(finalAdditions, additions[i])
	}

	if len(finalAdditions) == 0 {
		return
	}

	sort.Sort(finalAdditions)

	for i := range finalAdditions {
		e.WriteByte('\n')
		e.WriteExpr(finalAdditions[i])
	}
}

func (additions Additions) Len() int {
	return len(additions)
}

func (additions Additions) Less(i, j int) bool {
	return additions[i].AdditionType() < additions[j].AdditionType()
}

func (additions Additions) Swap(i, j int) {
	additions[i], additions[j] = additions[j], additions[i]
}

func AsAddition(expr SqlExpr) *OtherAddition {
	return &OtherAddition{
		SqlExpr: expr,
	}
}

type OtherAddition struct {
	SqlExpr
}

func (OtherAddition) AdditionType() AdditionType {
	return AdditionOther
}

func (a *OtherAddition) IsNil() bool {
	return a == nil || IsNilExpr(a.SqlExpr)
}
