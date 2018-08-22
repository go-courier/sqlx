package builder

func Select(sqlExpr SqlExpr, modifiers ...string) *StmtSelect {
	return &StmtSelect{
		sqlExpr:   sqlExpr,
		modifiers: modifiers,
	}
}

// https://dev.mysql.com/doc/refman/5.7/en/select.html
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

func (s *StmtSelect) Expr() *Expression {
	selectSql := "SELECT"

	if len(s.modifiers) > 0 {
		for i := range s.modifiers {
			selectSql += " " + s.modifiers[i]
		}
	}

	if s.sqlExpr == nil {
		s.sqlExpr = Expr("*")
	}

	expr := MustJoinExpr(" ", Expr(selectSql), s.sqlExpr)

	if s.table == nil {
		panic("Select should call method `From` to bind table")
	}

	expr = MustJoinExpr(" FROM ", expr, s.table)

	if len(s.additions) > 0 {
		expr = MustJoinExpr(" ", expr, s.additions)
	}

	return expr
}

func ForUpdate() *otherAddition {
	return (*otherAddition)(Expr("FOR UPDATE"))
}

func LockInShareMode() *otherAddition {
	return (*otherAddition)(Expr("LOCK IN SHARE MODE"))
}
