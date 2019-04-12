package builder

func Count(sqlExprs ...SqlExpr) *Function {
	if len(sqlExprs) == 0 {
		return Func("COUNT", Expr("1"))
	}
	return Func("COUNT", sqlExprs...)
}

func Avg(sqlExprs ...SqlExpr) *Function {
	return Func("AVG", sqlExprs...)
}

func Distinct(sqlExprs ...SqlExpr) *Function {
	return Func("DISTINCT", sqlExprs...)
}

func Min(sqlExprs ...SqlExpr) *Function {
	return Func("MIN", sqlExprs...)
}

func Max(sqlExprs ...SqlExpr) *Function {
	return Func("MAX", sqlExprs...)
}

func First(sqlExprs ...SqlExpr) *Function {
	return Func("FIRST", sqlExprs...)
}

func Last(sqlExprs ...SqlExpr) *Function {
	return Func("LAST", sqlExprs...)
}

func Sum(sqlExprs ...SqlExpr) *Function {
	return Func("SUM", sqlExprs...)
}

func Func(name string, sqlExprs ...SqlExpr) *Function {
	if name == "" {
		return nil
	}
	return &Function{
		Name:  name,
		Exprs: sqlExprs,
	}
}

type Function struct {
	Name  string
	Exprs []SqlExpr
}

func (f *Function) IsNil() bool {
	return f == nil || f.Name == ""
}

func (f *Function) Expr() *Ex {
	if f == nil {
		return nil
	}

	if f.Exprs != nil {
		e := Expr(f.Name)
		e.WriteGroup(func(e *Ex) {

		})
	}
	e := Expr(f.Name)
	e.WriteGroup(func(e *Ex) {
		if len(f.Exprs) == 0 {
			e.WriteByte('*')
		}
		for i := range f.Exprs {
			if i > 0 {
				e.WriteByte(',')
			}
			e.WriteExpr(f.Exprs[i])
		}
	})
	return e
}
