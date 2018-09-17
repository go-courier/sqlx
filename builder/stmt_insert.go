package builder

func Insert(modifiers ...string) *StmtInsert {
	return &StmtInsert{
		modifiers: modifiers,
	}
}

// https://dev.mysql.com/doc/refman/5.6/en/insert.html
type StmtInsert struct {
	table       *Table
	modifiers   []string
	assignments Assignments
	bySet       bool
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

func (s StmtInsert) Set(assignments ...*Assignment) *StmtInsert {
	s.assignments = Assignments(assignments)
	s.bySet = true
	return &s
}

func (s *StmtInsert) Expr() *Expression {
	selectSql := "INSERT"

	if len(s.modifiers) > 0 {
		for i := range s.modifiers {
			selectSql += " " + s.modifiers[i]
		}
	}

	if s.table == nil {
		panic("INSERT should bind table, please call INTO()")
	}

	expr := MustJoinExpr(" INTO ", Expr(selectSql), s.table)

	if len(s.assignments) == 0 {
		panic("INSERT should contain assignments, please call Set() or Values()")
	}

	if s.bySet {
		expr = MustJoinExpr(" SET ", expr, s.assignments)
	} else {
		expr = MustJoinExpr(" ", expr, s.assignments)
	}

	if len(s.additions) > 0 {
		expr = MustJoinExpr(" ", expr, s.additions)
	}

	return expr
}

func OnDuplicateKeyUpdate(assignments ...*Assignment) *otherAddition {
	return (*otherAddition)(MustJoinExpr("ON DUPLICATE KEY UPDATE ", Expr(""), Assignments(assignments)))
}
