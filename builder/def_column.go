package builder

import (
	"context"
	"reflect"
	"strings"

	"github.com/go-courier/x/types"
)

func Col(name string) *Column {
	return &Column{
		Name:       strings.ToLower(name),
		ColumnType: &ColumnType{},
	}
}

var _ TableDefinition = (*Column)(nil)

type Column struct {
	Name      string
	FieldName string
	Table     *Table
	exactly   bool

	*ColumnType
}

func (c Column) Full() *Column {
	c.exactly = true
	return &c
}

func (c *Column) Of(table *Table) *Column {
	col := &Column{
		Name:       c.Name,
		FieldName:  c.FieldName,
		Table:      table,
		exactly:    true,
		ColumnType: c.ColumnType,
	}
	return col
}

func (c *Column) IsNil() bool {
	return c == nil
}

func (c *Column) Ex(ctx context.Context) *Ex {
	toggles := TogglesFromContext(ctx)
	if c.Table != nil && (c.exactly || toggles.Is(ToggleMultiTable)) {
		if toggles.Is(ToggleNeedAutoAlias) {
			return Expr("?.? AS ?", c.Table, Expr(c.Name), Expr(c.Name)).Ex(ctx)
		}
		return Expr("?.?", c.Table, Expr(c.Name)).Ex(ctx)
	}
	return ExactlyExpr(c.Name).Ex(ctx)
}

func (c *Column) Expr(query string, args ...interface{}) *Ex {
	n := len(args)
	e := Expr("")
	e.Grow(n)

	qc := 0

	for _, key := range []byte(query) {
		switch key {
		case '#':
			e.WriteExpr(c)
		case '?':
			e.WriteQueryByte(key)
			if n > qc {
				e.AppendArgs(args[qc])
				qc++
			}
		default:
			e.WriteQueryByte(key)
		}
	}

	return e
}

func (c Column) Field(fieldName string) *Column {
	c.FieldName = fieldName
	return &c
}

func (c Column) Type(v interface{}, tagValue string) *Column {
	c.ColumnType = ColumnTypeFromTypeAndTag(types.FromRType(reflect.TypeOf(v)), tagValue)
	return &c
}

func (c Column) On(table *Table) *Column {
	c.Table = table
	return &c
}

func (c *Column) T() *Table {
	return c.Table
}

func (c *Column) ValueBy(v interface{}) *Assignment {
	return ColumnsAndValues(c, v)
}

func (c *Column) Incr(d int) SqlExpr {
	return Expr("? + ?", c, d)
}

func (c *Column) Desc(d int) SqlExpr {
	return Expr("? - ?", c, d)
}

func (c *Column) Like(v string) SqlCondition {
	return AsCond(Expr("? LIKE ?", c, "%"+v+"%"))
}

func (c *Column) LeftLike(v string) SqlCondition {
	return AsCond(Expr("? LIKE ?", c, "%"+v))
}

func (c *Column) RightLike(v string) SqlCondition {
	return AsCond(Expr("? LIKE ?", c, v+"%"))
}

func (c *Column) NotLike(v string) SqlCondition {
	return AsCond(Expr("? NOT LIKE ?", c, "%"+v+"%"))
}

func (c *Column) IsNull() SqlCondition {
	return AsCond(Expr("? IS NULL", c))
}

func (c *Column) IsNotNull() SqlCondition {
	return AsCond(Expr("? IS NOT NULL", c))
}

func (c *Column) Between(leftValue interface{}, rightValue interface{}) SqlCondition {
	return AsCond(Expr("? BETWEEN ? AND ?", c, leftValue, rightValue))
}

func (c *Column) NotBetween(leftValue interface{}, rightValue interface{}) SqlCondition {
	return AsCond(Expr("? NOT BETWEEN ? AND ?", c, leftValue, rightValue))
}

type WithConditionFor interface {
	ConditionFor(c *Column) SqlCondition
}

func (c *Column) In(args ...interface{}) SqlCondition {
	n := len(args)

	switch n {
	case 0:
		return nil
	case 1:
		if withConditionFor, ok := args[0].(WithConditionFor); ok {
			return withConditionFor.ConditionFor(c)
		}
	}

	e := Expr("? IN ")

	e.Grow(n + 1)

	e.AppendArgs(c)

	e.WriteGroup(func(e *Ex) {
		for i := 0; i < n; i++ {
			e.WriteHolder(i)
		}
	})

	e.AppendArgs(args...)

	return AsCond(e)
}

func (c *Column) NotIn(args ...interface{}) SqlCondition {
	n := len(args)
	if n == 0 {
		return nil
	}

	e := Expr("")
	e.Grow(n + 1)

	e.WriteQuery("? NOT IN ")
	e.AppendArgs(c)

	e.WriteGroup(func(e *Ex) {
		for i := 0; i < n; i++ {
			e.WriteHolder(i)
		}
	})

	e.AppendArgs(args...)

	return AsCond(e)
}

func (c *Column) Eq(v interface{}) SqlCondition {
	return AsCond(Expr("? = ?", c, v))
}

func (c *Column) Neq(v interface{}) SqlCondition {
	return AsCond(Expr("? <> ?", c, v))
}

func (c *Column) Gt(v interface{}) SqlCondition {
	return AsCond(Expr("? > ?", c, v))
}

func (c *Column) Gte(v interface{}) SqlCondition {
	return AsCond(Expr("? >= ?", c, v))
}

func (c *Column) Lt(v interface{}) SqlCondition {
	return AsCond(Expr("? < ?", c, v))
}

func (c *Column) Lte(v interface{}) SqlCondition {
	return AsCond(Expr("? <= ?", c, v))
}
