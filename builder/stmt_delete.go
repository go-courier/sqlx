package builder

func Delete(modifiers ...string) *StmtDelete {
	return &StmtDelete{
		modifiers: modifiers,
	}
}

type StmtDelete struct {
	table     *Table
	modifiers []string
	additions Additions
}

func (s StmtDelete) From(table *Table, additions ...Addition) *StmtDelete {
	s.table = table
	s.additions = additions
	return &s
}

func (s *StmtDelete) Expr() *Expression {
	selectSql := "DELETE"

	if len(s.modifiers) > 0 {
		for i := range s.modifiers {
			selectSql += " " + s.modifiers[i]
		}
	}

	expr := Expr(selectSql)

	if s.table == nil {
		panic("DELETE should call method `From` to bind table")
	}

	expr = MustJoinExpr(" FROM ", expr, s.table)

	if len(s.additions) > 0 {
		expr = MustJoinExpr(" ", expr, s.additions)
	}

	return expr
}
