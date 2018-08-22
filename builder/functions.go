package builder

func Count(sqlExprs ...SqlExpr) *Function {
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
	if len(sqlExprs) == 0 {
		return &Function{
			Name: name,
			expr: Expr("*"),
		}
	}
	return &Function{
		Name: name,
		expr: MustJoinExpr(", ", sqlExprs...),
	}
}

type Function struct {
	Name string
	expr SqlExpr
}

func (f *Function) Expr() *Expression {
	if f == nil {
		return nil
	}
	if f.expr != nil {
		e := f.expr.Expr()
		if e != nil {
			return Expr(f.Name+"("+e.Query+")", e.Args...)
		}
	}
	return nil
}
