package builder

import (
	"context"
)

type SelectStatement interface {
	SqlExpr
	selectStatement()
}

func Select(sqlExpr SqlExpr, modifiers ...string) *StmtSelect {
	return &StmtSelect{
		sqlExpr:   sqlExpr,
		modifiers: modifiers,
	}
}

type StmtSelect struct {
	SelectStatement
	sqlExpr   SqlExpr
	table     *Table
	modifiers []string
	additions []Addition
}

func (s *StmtSelect) IsNil() bool {
	return s == nil
}

func (s StmtSelect) From(table *Table, additions ...Addition) *StmtSelect {
	s.table = table
	s.additions = additions
	return &s
}

func (s *StmtSelect) Ex(ctx context.Context) *Ex {
	multiTable := false

	for i := range s.additions {
		addition := s.additions[i]
		if IsNilExpr(addition) {
			continue
		}

		if addition.AdditionType() == AdditionJoin {
			multiTable = true
		}
	}

	if multiTable {
		ctx = ContextWithToggles(ctx, Toggles{
			ToggleMultiTable: multiTable,
		})
	}

	e := Expr("SELECT")

	if len(s.modifiers) > 0 {
		for i := range s.modifiers {
			e.WriteByte(' ')
			e.WriteString(s.modifiers[i])
		}
	}

	sqlExpr := s.sqlExpr

	if IsNilExpr(sqlExpr) {
		sqlExpr = Expr("*")
	}

	e.WriteByte(' ')
	e.WriteExpr(sqlExpr)

	if !IsNilExpr(s.table) {
		e.WriteString(" FROM ")
		e.WriteExpr(s.table)
	}

	WriteAdditions(e, s.additions...)

	return e.Ex(ctx)
}

func ForUpdate() *OtherAddition {
	return AsAddition(Expr("FOR UPDATE"))
}
