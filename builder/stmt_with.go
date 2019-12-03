package builder

import (
	"context"
	"strings"
)

func WithRecursive(t *Table) *WithStmt {
	return With(t, "RECURSIVE")
}

func With(t *Table, modifiers ...string) *WithStmt {
	return &WithStmt{
		t:         t,
		modifiers: modifiers,
	}
}

type WithStmt struct {
	t         *Table
	modifiers []string
	as        func(t *Table) SqlExpr
	do        func(t *Table) SqlExpr
}

func (w *WithStmt) IsNil() bool {
	return w == nil || w.as == nil || w.do == nil
}

func (w *WithStmt) T() *Table {
	return w.t
}

func (w WithStmt) As(as func(t *Table) SqlExpr) *WithStmt {
	w.as = as
	return &w
}

func (w WithStmt) Do(do func(t *Table) SqlExpr) *WithStmt {
	w.do = do
	return &w
}

func (w *WithStmt) Ex(ctx context.Context) *Ex {
	e := Expr("WITH ")

	if len(w.modifiers) > 0 {
		e.WriteString(strings.Join(w.modifiers, " "))
		e.WriteString(" ")
	}

	t := w.T()

	e.WriteExpr(w.T())
	e.WriteGroup(func(e *Ex) {
		e.WriteExpr(&t.Columns)
	})

	e.WriteString(" AS ")

	e.WriteGroup(func(e *Ex) {
		e.WriteByte('\n')
		e.WriteExpr(w.as(t))
		e.WriteByte('\n')
	})

	e.WriteByte('\n')
	e.WriteExpr(w.do(t))

	return e.Ex(ctx)
}
