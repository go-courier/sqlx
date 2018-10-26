package builder

import (
	"errors"
)

var (
	UpdateNeedLimitByWhere = errors.New("no where limit for update")
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

func (s *StmtUpdate) IsNil() bool {
	return s == nil || s.table == nil || s.assignments.IsNil()
}

func (s *StmtUpdate) Expr() *Ex {
	if s.IsNil() {
		return nil
	}

	e := Expr("UPDATE")

	if len(s.modifiers) > 0 {
		for i := range s.modifiers {
			e.WriteByte(' ')
			e.WriteString(s.modifiers[i])
		}
	}

	e.WriteByte(' ')
	e.WriteExpr(s.table)
	e.WriteString(" SET ")
	e.WriteExpr(s.assignments)

	if !s.additions.IsNil() {
		e.WriteExpr(s.additions)
	}

	return e
}
