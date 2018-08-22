package builder

import (
	"strings"
)

type SqlExpr interface {
	Expr() *Expression
}

func Expr(query string, args ...interface{}) *Expression {
	return &Expression{Query: query, Args: args}
}

func ExprFrom(v interface{}) *Expression {
	if v == nil {
		return nil
	}
	switch v.(type) {
	case *Expression:
		return v.(*Expression)
	case SqlExpr:
		return v.(SqlExpr).Expr()
	}
	return nil
}

func MustJoinExpr(joiner string, sqlExprs ...SqlExpr) *Expression {
	e, err := JoinExpr(joiner, sqlExprs...)
	if err != nil {
		panic(err)
	}
	return e
}

func JoinExpr(joiner string, sqlExprs ...SqlExpr) (*Expression, error) {
	query := ""
	var args []interface{}

	for i, sqlExpr := range sqlExprs {
		if sqlExpr == nil {
			continue
		}

		e := sqlExpr.Expr()
		if e.Err != nil {
			return nil, e.Err
		}
		if i > 0 {
			query += joiner
		}
		query += e.Query
		args = append(args, e.Args...)
	}

	return Expr(query, args...), nil
}

func ExprErr(err error) *Expression {
	return &Expression{Err: err}
}

type Expression struct {
	Query string
	Args  []interface{}
	Err   error
}

func (e *Expression) Expr() *Expression {
	return e
}

func HolderRepeat(length int) string {
	if length > 1 {
		return strings.Repeat("?,", length-1) + "?"
	}
	return "?"
}

func quote(n string) string {
	return "`" + n + "`"
}
