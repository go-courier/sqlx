package builder

func Select(sqlExpr SqlExpr, modifiers ...string) *StmtSelect {
	return &StmtSelect{
		sqlExpr:   sqlExpr,
		modifiers: modifiers,
	}
}

type StmtSelect struct {
	sqlExpr   SqlExpr
	table     *Table
	modifiers []string
	additions Additions
}

func (s StmtSelect) From(table *Table, additions ...Addition) *StmtSelect {
	s.table = table
	s.additions = additions
	return &s
}

func (s *StmtSelect) IsNil() bool {
	return s == nil || s.table == nil
}

func (s *StmtSelect) Expr() *Ex {
	e := Expr("SELECT")

	if len(s.modifiers) > 0 {
		for i := range s.modifiers {
			e.WriteByte(' ')
			e.WriteString(s.modifiers[i])
		}
	}

	if s.sqlExpr == nil {
		s.sqlExpr = Expr("*")
	}

	e.WriteByte(' ')
	e.WriteExpr(s.sqlExpr)

	e.WriteString(" FROM ")
	e.WriteExpr(s.table)

	if !s.additions.IsNil() {
		e.WriteExpr(s.additions)
	}

	return e
}

func ForUpdate() *otherAddition {
	return AsAddition(Expr("FOR UPDATE"))
}
