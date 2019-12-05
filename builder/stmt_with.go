package builder

import (
	"context"
	"strings"
)

type BuildSubQuery func(table *Table) SqlExpr

func WithRecursive(t *Table, build BuildSubQuery) *WithStmt {
	return With(t, build, "RECURSIVE")
}

func With(t *Table, build BuildSubQuery, modifiers ...string) *WithStmt {
	return (&WithStmt{modifiers: modifiers}).With(t, build)
}

type WithStmt struct {
	modifiers []string
	tables    []*Table
	asList    []BuildSubQuery
	statement func(tables ...*Table) SqlExpr
}

func (w *WithStmt) IsNil() bool {
	return w == nil || len(w.tables) == 0 || len(w.asList) == 0 || w.statement == nil
}

func (w WithStmt) With(t *Table, build BuildSubQuery) *WithStmt {
	w.tables = append(w.tables, t)
	w.asList = append(w.asList, build)
	return &w
}

func (w WithStmt) Exec(statement func(tables ...*Table) SqlExpr) *WithStmt {
	w.statement = statement
	return &w
}

func (w *WithStmt) Ex(ctx context.Context) *Ex {
	e := Expr("WITH ")

	if len(w.modifiers) > 0 {
		e.WriteString(strings.Join(w.modifiers, " "))
		e.WriteString(" ")
	}

	for i := range w.tables {
		if i > 0 {
			e.WriteString(", ")
		}

		table := w.tables[i]

		e.WriteExpr(table)
		e.WriteGroup(func(e *Ex) {
			e.WriteExpr(&table.Columns)
		})

		e.WriteString(" AS ")

		build := w.asList[i]

		e.WriteGroup(func(e *Ex) {
			e.WriteByte('\n')
			e.WriteExpr(build(table))
			e.WriteByte('\n')
		})
	}

	e.WriteByte('\n')
	e.WriteExpr(w.statement(w.tables...))

	return e.Ex(ctx)
}
