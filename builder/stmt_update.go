package builder

import (
	"context"
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
	assignments []*Assignment
	additions   []Addition
}

func (s *StmtUpdate) IsNil() bool {
	return s == nil || IsNilExpr(s.table) || len(s.assignments) == 0
}

func (s StmtUpdate) Set(assignments ...*Assignment) *StmtUpdate {
	s.assignments = assignments
	return &s
}

func (s StmtUpdate) Where(c SqlCondition, additions ...Addition) *StmtUpdate {
	s.additions = []Addition{Where(c)}
	if len(additions) > 0 {
		s.additions = append(s.additions, additions...)
	}
	return &s
}

func (s *StmtUpdate) Ex(ctx context.Context) *Ex {
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

	WriteAssignments(e, s.assignments...)
	WriteAdditions(e, s.additions...)

	return e.Ex(ctx)
}
