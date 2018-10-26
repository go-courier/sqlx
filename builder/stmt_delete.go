package builder

func Delete() *StmtDelete {
	return &StmtDelete{}
}

type StmtDelete struct {
	table     *Table
	additions Additions
}

func (s StmtDelete) From(table *Table, additions ...Addition) *StmtDelete {
	s.table = table
	s.additions = additions
	return &s
}

func (s *StmtDelete) IsNil() bool {
	return s == nil || s.table == nil
}

func (s *StmtDelete) Expr() *Ex {
	if s.IsNil() {
		return nil
	}

	expr := Expr("DELETE FROM ")
	expr.WriteExpr(s.table)

	if !s.additions.IsNil() {
		expr.WriteExpr(s.additions)
	}

	return expr
}
