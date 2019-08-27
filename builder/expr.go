package builder

import (
	"bytes"
	"database/sql/driver"
	"reflect"

	"github.com/go-courier/reflectx"
)

type SqlExpr interface {
	IsNil() bool
	Expr() *Ex
}

func ExprFrom(v interface{}) *Ex {
	if v == nil {
		return nil
	}
	switch e := v.(type) {
	case *Ex:
		return e
	case SqlExpr:
		if e.IsNil() {
			return nil
		}
		return e.Expr()
	}
	return nil
}

func MultiExpr(exprs ...SqlExpr) SqlExpr {
	e := Expr("")
	for i := range exprs {
		if i != 0 {
			e.WriteString(", ")
		}
		e.WriteExpr(exprs[i])
	}
	return e
}

func Expr(query string, args ...interface{}) *Ex {
	return &Ex{Buffer: bytes.NewBufferString(query), args: args}
}

func Alias(expr SqlExpr, name string) *AliasEx {
	return &AliasEx{
		Name:    name,
		SqlExpr: expr,
	}
}

type AliasEx struct {
	Name string
	SqlExpr
}

func (expr *AliasEx) Expr() *Ex {
	e := Expr("(?) AS ?")
	e.AppendArgs(expr.SqlExpr, Expr(expr.Name))
	return e
}

type Ex struct {
	*bytes.Buffer
	args []interface{}
	err  error
}

func (e *Ex) IsNil() bool {
	return e == nil || e.Len() == 0
}

func (e *Ex) Query() string {
	return e.String()
}

func (e *Ex) Args() []interface{} {
	return e.args
}

func (e *Ex) Err() error {
	return e.err
}

func (e *Ex) AppendArgs(args ...interface{}) {
	e.args = append(e.args, args...)
}

func (e *Ex) ArgsLen() int {
	return len(e.args)
}

func (e *Ex) Expr() *Ex {
	if e.IsNil() {
		return nil
	}
	return e
}

func (e *Ex) WriteGroup(fn func(e *Ex)) {
	e.WriteByte('(')
	fn(e)
	e.WriteByte(')')
}

func (e *Ex) WhiteComments(comments string) {
	e.WriteString("/* ")
	e.WriteString(comments)
	e.WriteString(" */")
}

func (e *Ex) WriteExpr(expr SqlExpr) {
	ex := ExprFrom(expr)
	if ex == nil {
		return
	}
	e.Write(ex.Bytes())
	e.AppendArgs(ex.Args()...)
}

func (e *Ex) WriteEnd() {
	e.WriteByte(';')
}

func (e *Ex) WriteHolder(idx int) {
	if idx > 0 {
		e.WriteByte(',')
	}
	e.WriteByte('?')
}

func (e *Ex) Flatten() *Ex {
	index := 0
	expr := Expr("")
	data := e.Bytes()
	argsLen := e.ArgsLen()

	for i := range data {
		c := data[i]
		switch c {
		case '?':
			if argsLen == 0 {
				expr.WriteByte(c)
				continue
			}
			arg := e.args[index]
			switch a := arg.(type) {
			case ValuerExpr:
				expr.WriteString(a.ValueEx())
				expr.AppendArgs(arg)
			case SqlExpr:
				e := ExprFrom(a)
				if !e.IsNil() {
					expr.WriteExpr(e.Flatten())
				}
			case driver.Valuer:
				expr.WriteHolder(0)
				expr.AppendArgs(arg)
			default:
				typ := reflect.TypeOf(arg)
				if !reflectx.IsBytes(typ) && typ.Kind() == reflect.Slice {
					sliceRv := reflect.ValueOf(arg)
					length := sliceRv.Len()
					for i := 0; i < length; i++ {
						expr.WriteHolder(i)
						expr.AppendArgs(sliceRv.Index(i).Interface())
					}
				} else {
					expr.WriteHolder(0)
					expr.AppendArgs(arg)
				}
			}
			index++
		default:
			expr.WriteByte(c)
		}
	}

	return expr
}

func (e *Ex) ReplaceValueHolder(bindVar func(idx int) string) *Ex {
	index := 0
	expr := Expr("")
	data := e.Bytes()
	argsLen := e.ArgsLen()

	for i := range data {
		c := data[i]
		switch c {
		case '?':
			if argsLen == 0 {
				expr.WriteByte(c)
				continue
			}
			expr.WriteString(bindVar(index))
			expr.AppendArgs(e.args[index])
			index++
		default:
			expr.WriteByte(c)
		}
	}

	return expr
}
