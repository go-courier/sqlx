package builder

import (
	"context"
)

func Delete() *StmtDelete {
	return &StmtDelete{}
}

type StmtDelete struct {
	table     *Table
	additions []Addition
}

func (s *StmtDelete) IsNil() bool {
	return s == nil || IsNilExpr(s.table)
}

func (s StmtDelete) From(table *Table, additions ...Addition) *StmtDelete {
	s.table = table
	s.additions = additions
	return &s
}

func (s *StmtDelete) Ex(ctx context.Context) *Ex {
	e := Expr("DELETE FROM ")

	e.WriteExpr(s.table)

	WriteAdditions(e, s.additions...)

	return e.Ex(ctx)
}
