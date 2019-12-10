package builder

import (
	"context"
)

func Insert(modifiers ...string) *StmtInsert {
	return &StmtInsert{
		modifiers: modifiers,
	}
}

// https://dev.mysql.com/doc/refman/5.6/en/insert.html
type StmtInsert struct {
	table       *Table
	modifiers   []string
	assignments []*Assignment
	additions   Additions
}

func (s StmtInsert) Into(table *Table, additions ...Addition) *StmtInsert {
	s.table = table
	s.additions = additions
	return &s
}

func (s StmtInsert) Values(cols *Columns, values ...interface{}) *StmtInsert {
	s.assignments = Assignments{ColumnsAndValues(cols, values...)}
	return &s
}

func (s *StmtInsert) IsNil() bool {
	return s == nil || s.table == nil || len(s.assignments) == 0
}

func (s *StmtInsert) Ex(ctx context.Context) *Ex {
	e := Expr("INSERT")

	if len(s.modifiers) > 0 {
		for i := range s.modifiers {
			e.WriteByte(' ')
			e.WriteString(s.modifiers[i])
		}
	}

	e.WriteString(" INTO ")
	e.WriteExpr(s.table)
	e.WriteByte(' ')

	e.WriteExpr(ExprBy(func(ctx context.Context) *Ex {
		e := Expr("")

		ctx = ContextWithToggles(ctx, Toggles{
			ToggleUseValues: true,
		})

		WriteAssignments(e, s.assignments...)

		return e.Ex(ctx)
	}))

	WriteAdditions(e, s.additions...)

	return e.Ex(ctx)
}

func OnDuplicateKeyUpdate(assignments ...*Assignment) *OtherAddition {
	assigns := assignments
	if len(assignments) == 0 {
		return nil
	}

	e := Expr("ON DUPLICATE KEY UPDATE ")

	for i := range assigns {
		if i > 0 {
			e.WriteString(", ")
		}
		e.WriteExpr(assigns[i])
	}

	return AsAddition(e)
}

func Returning(expr SqlExpr) *OtherAddition {
	e := Expr("RETURNING ")
	if expr == nil || expr.IsNil() {
		e.WriteByte('*')
	} else {
		e.WriteExpr(expr)
	}
	return AsAddition(e)
}
