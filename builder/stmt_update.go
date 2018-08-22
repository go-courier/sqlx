package builder

import (
	"fmt"
)

var (
	UpdateNeedLimitByWhere = fmt.Errorf("no where limit for update")
)

func Update(table *Table, modifiers ...string) *StmtUpdate {
	return &StmtUpdate{
		table:     table,
		modifiers: modifiers,
	}
}

type StmtUpdate struct {
	table       *Table
	modifiers   []string
	assignments Assignments
	additions   Additions
}

func (s StmtUpdate) Set(assignments ...*Assignment) *StmtUpdate {
	s.assignments = Assignments(assignments)
	return &s
}

func (s StmtUpdate) Where(c *Condition, additions ...Addition) *StmtUpdate {
	s.additions = []Addition{Where(c)}
	if len(additions) > 0 {
		s.additions = append(s.additions, additions...)
	}
	return &s
}

func (s *StmtUpdate) Expr() *Expression {
	selectSql := "UPDATE"

	if len(s.modifiers) > 0 {
		for i := range s.modifiers {
			selectSql += " " + s.modifiers[i]
		}
	}

	expr := Expr(selectSql)

	if s.table == nil {
		panic("UPDATE should bind table")
	}

	if len(s.assignments) == 0 {
		panic("UPDATE should contain assignments, please call Set() to set it")
	}

	expr = MustJoinExpr(" SET ", MustJoinExpr(" ", expr, s.table), s.assignments)

	if len(s.additions) > 0 {
		expr = MustJoinExpr(" ", expr, s.additions)
	}

	return expr
}
