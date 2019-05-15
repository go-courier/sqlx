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
	return s == nil || s.table == nil || s.assignments.IsNil()
}

func (s *StmtInsert) Expr() *Ex {
	if s.IsNil() {
		return nil
	}

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
	e.WriteExpr(s.assignments)

	if len(s.additions) > 0 {
		e.WriteExpr(s.additions)
	}

	return e
}

func OnDuplicateKeyUpdate(assignments ...*Assignment) *OtherAddition {
	assigns := Assignments(assignments)
	if assigns.IsNil() {
		return nil
	}

	e := Expr("ON DUPLICATE KEY UPDATE ")
	e.WriteExpr(assigns)
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
