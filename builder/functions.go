package builder

import (
	"context"
)

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
		name:  name,
		exprs: sqlExprs,
	}
}

type Function struct {
	name  string
	exprs []SqlExpr
}

func (f *Function) IsNil() bool {
	return f == nil || f.name == ""
}

func (f *Function) Ex(ctx context.Context) *Ex {
	e := Expr(f.name)

	e.WriteGroup(func(e *Ex) {
		if len(f.exprs) == 0 {
			e.WriteByte('*')
		}

		for i := range f.exprs {
			if i > 0 {
				e.WriteByte(',')
			}
			e.WriteExpr(f.exprs[i])
		}
	})

	return e.Ex(ctx)
}
