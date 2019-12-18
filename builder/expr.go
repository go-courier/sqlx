package builder

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"

	"github.com/go-courier/reflectx"
)

type SqlExpr interface {
	IsNil() bool
	Ex(ctx context.Context) *Ex
}

func IsNilExpr(e SqlExpr) bool {
	return e == nil || e.IsNil()
}

func RangeNotNilExpr(exprs []SqlExpr, each func(e SqlExpr, i int)) {
	count := 0

	for i := range exprs {
		e := exprs[i]
		if IsNilExpr(e) {
			continue
		}
		each(e, count)
		count++
	}
}

func Expr(query string, args ...interface{}) *Ex {
	return &Ex{Buffer: bytes.NewBufferString(query), args: args}
}

func ResolveExpr(v interface{}) *Ex {
	return ResolveExprContext(context.Background(), v)
}

func ResolveExprContext(ctx context.Context, v interface{}) *Ex {
	switch e := v.(type) {
	case nil:
		return nil
	case SqlExpr:
		if IsNilExpr(e) {
			return nil
		}
		return e.Ex(ctx)
	}
	return nil
}

func Multi(exprs ...SqlExpr) SqlExpr {
	return MultiWith(" ", exprs...)
}

func MultiWith(connector string, exprs ...SqlExpr) SqlExpr {
	return ExprBy(func(ctx context.Context) *Ex {
		e := Expr("")
		for i := range exprs {
			if i != 0 {
				e.WriteString(connector)
			}
			e.WriteExpr(exprs[i])
		}
		return e.Ex(ctx)
	})
}

func ExprBy(build func(ctx context.Context) *Ex) SqlExpr {
	return &exBy{build: build}
}

type exBy struct {
	build func(ctx context.Context) *Ex
}

func (c *exBy) IsNil() bool {
	return c == nil || c.build == nil
}

func (c *exBy) Ex(ctx context.Context) *Ex {
	return c.build(ctx)
}

type Ex struct {
	*bytes.Buffer
	args     []interface{}
	err      error
	rendered bool
	ident    int
}

func (e *Ex) IsNil() bool {
	return e == nil || e.Len() == 0
}

func (e *Ex) Query() string {
	if e == nil {
		return ""
	}
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

func (e Ex) Ex(ctx context.Context) *Ex {
	if e.rendered {
		return &e
	}

	if e.IsNil() {
		return nil
	}

	if e.ArgsLen() == 0 {
		e.rendered = true
		return &e
	}

	index := 0
	expr := Expr("")
	expr.rendered = true

	query := e.Bytes()
	n := len(e.args)

	for i := range query {
		c := query[i]
		switch c {
		case '?':
			if index >= n {
				panic(fmt.Errorf("missing arg %d of %s", index, query))
			}

			arg := e.args[index]

			switch a := arg.(type) {
			case ValuerExpr:
				expr.WriteString(a.ValueEx())
				expr.AppendArgs(arg)
			case SqlExpr:
				if !IsNilExpr(a) {
					subEx := a.Ex(ctx)
					if !IsNilExpr(subEx) {
						expr.Write(subEx.Bytes())
						expr.AppendArgs(subEx.Args()...)
					}
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

func (e *Ex) WriteGroup(fn func(e *Ex)) {
	e.WriteByte('(')
	fn(e)
	e.WriteByte(')')
}

func (e *Ex) WhiteComments(comments []byte) {
	e.WriteString("/* ")
	e.Write(comments)
	e.WriteString(" */")
}

func (e *Ex) WriteExpr(expr SqlExpr) {
	if IsNilExpr(expr) {
		return
	}

	e.WriteHolder(0)
	e.AppendArgs(expr)
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
