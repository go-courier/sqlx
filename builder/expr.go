package builder

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"

	reflectx "github.com/go-courier/x/reflect"

	contextx "github.com/go-courier/x/context"
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

func ExactlyExpr(query string, args ...interface{}) *Ex {
	if query != "" {
		return &Ex{b: *bytes.NewBufferString(query), args: args, exactly: true}
	}
	return &Ex{args: args, exactly: true}
}

func Expr(query string, args ...interface{}) *Ex {
	if query != "" {
		return &Ex{b: *bytes.NewBufferString(query), args: args}
	}
	return &Ex{args: args}
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
		e.Grow(len(exprs))

		for i := range exprs {
			if i != 0 {
				e.WriteQuery(connector)
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
	b       bytes.Buffer
	args    []interface{}
	err     error
	exactly bool
}

func (e *Ex) IsNil() bool {
	return e == nil || e.b.Len() == 0
}

func (e *Ex) Query() string {
	if e == nil {
		return ""
	}
	return e.b.String()
}

func (e *Ex) Args() []interface{} {
	if len(e.args) == 0 {
		return nil
	}
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

func (e *Ex) WriteString(s string) (int, error) {
	return e.b.WriteString(s)
}

func (e *Ex) WriteByte(b byte) error {
	return e.b.WriteByte(b)
}

func (e *Ex) QueryGrow(n int) {
	e.b.Grow(n)
}

func (e *Ex) Grow(n int) {
	if n > 0 && cap(e.args)-len(e.args) < n {
		args := make([]interface{}, len(e.args), 2*cap(e.args)+n)
		copy(args, e.args)
		e.args = args
	}
}

func (e *Ex) WriteQuery(s string) {
	_, _ = e.b.WriteString(s)
}

func (e *Ex) WriteQueryByte(b byte) {
	_ = e.b.WriteByte(b)
}

func (e *Ex) WriteGroup(fn func(e *Ex)) {
	e.WriteQueryByte('(')
	fn(e)
	e.WriteQueryByte(')')
}

func (e *Ex) WhiteComments(comments []byte) {
	_, _ = e.b.WriteString("/* ")
	_, _ = e.b.Write(comments)
	_, _ = e.b.WriteString(" */")
}

func (e *Ex) WriteExpr(expr SqlExpr) {
	if IsNilExpr(expr) {
		return
	}

	e.WriteHolder(0)
	e.AppendArgs(expr)
}

func (e *Ex) WriteEnd() {
	e.WriteQueryByte(';')
}

func (e *Ex) WriteHolder(idx int) {
	if idx > 0 {
		e.b.WriteByte(',')
	}
	e.b.WriteByte('?')
}

func (e *Ex) SetExactly(exactly bool) {
	e.exactly = exactly
}

func (e *Ex) Ex(ctx context.Context) *Ex {
	if e.IsNil() {
		return nil
	}

	args, n := e.args, len(e.args)

	eb := exprBuilderFromContext(ctx)
	eb.Grow(n)

	query := e.Query()

	if e.exactly {
		eb.WriteQuery(query)
		eb.AppendArgs(args...)
		eb.exactly = true
		return eb
	}

	shouldResolveArgs := preprocessArgs(args)

	if !shouldResolveArgs {
		eb.WriteQuery(query)
		eb.AppendArgs(args...)
		eb.SetExactly(true)
		return eb
	}

	argIndex := 0

	for i := range query {
		switch c := query[i]; c {
		case '?':
			if argIndex >= n {
				panic(fmt.Errorf("missing arg %d of %s", argIndex, query))
			}

			switch arg := args[argIndex].(type) {
			case SqlExpr:
				if !IsNilExpr(arg) {
					subExpr := arg.Ex(contextWithExprBuilder(ctx, eb))

					if subExpr != eb && !IsNilExpr(subExpr) {
						eb.WriteQuery(subExpr.Query())
						eb.AppendArgs(subExpr.Args()...)
					}
				}
			default:
				eb.WriteHolder(0)
				eb.AppendArgs(arg)
			}
			argIndex++
		default:
			eb.WriteQueryByte(c)
		}
	}

	eb.SetExactly(true)

	return eb
}

func exactlyExprFromSlice(values []interface{}) *Ex {
	if n := len(values); n > 0 {
		return ExactlyExpr(strings.Repeat(",?", n)[1:], values...)
	}
	return ExactlyExpr("")
}

func preprocessArgs(args []interface{}) bool {
	shouldResolve := false

	for i := range args {
		switch arg := args[i].(type) {
		case ValuerExpr:
			args[i] = ExactlyExpr(arg.ValueEx(), arg)
			shouldResolve = true
		case SqlExpr:
			shouldResolve = true
		case driver.Valuer:

		case []interface{}:
			args[i] = exactlyExprFromSlice(arg)
			shouldResolve = true
		default:
			if typ := reflect.TypeOf(arg); !reflectx.IsBytes(typ) && typ.Kind() == reflect.Slice {
				args[i] = exactlyExprFromSlice(toInterfaceSlice(arg))
				shouldResolve = true
			}
		}
	}

	return shouldResolve
}

func toInterfaceSlice(arg interface{}) []interface{} {
	switch x := (arg).(type) {
	case []bool:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []string:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []float32:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []float64:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []int:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []int8:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []int16:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []int32:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []int64:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []uint:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []uint8:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []uint16:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []uint32:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []uint64:
		values := make([]interface{}, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []interface{}:
		return x
	}
	sliceRv := reflect.ValueOf(arg)
	values := make([]interface{}, sliceRv.Len())
	for i := range values {
		values[i] = sliceRv.Index(i).Interface()
	}
	return values
}

type contextKeyEx struct{}

func contextWithExprBuilder(ctx context.Context, ex *Ex) context.Context {
	return contextx.WithValue(ctx, contextKeyEx{}, ex)
}

func exprBuilderFromContext(ctx context.Context) *Ex {
	if ex, ok := ctx.Value(contextKeyEx{}).(*Ex); ok {
		return ex
	}
	return Expr("")
}
