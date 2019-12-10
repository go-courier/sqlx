package builder

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-courier/reflectx"
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

	Description []string
	Relation    []string

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
			return Expr("(?.?) AS ?", c.Table, Expr(c.Name), Expr(c.Name)).Ex(ctx)
		}
		return Expr("?.?", c.Table, Expr(c.Name)).Ex(ctx)
	}
	return Expr(c.Name).Ex(ctx)
}

func (c *Column) Expr(query string, args ...interface{}) *Ex {
	e := Expr("")

	qc := 0
	n := len(args)

	for _, key := range []byte(query) {
		switch key {
		case '#':
			e.WriteExpr(c)
		case '?':
			e.WriteByte(key)
			if n > qc {
				e.AppendArgs(args[qc])
				qc++
			}
		default:
			e.WriteByte(key)
		}
	}

	return e
}

func (c Column) Field(fieldName string) *Column {
	c.FieldName = fieldName
	return &c
}

func (c Column) Type(v interface{}, tagValue string) *Column {
	c.ColumnType = ColumnTypeFromTypeAndTag(reflect.TypeOf(v), tagValue)
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
	return c.Expr("# + ?", d)
}

func (c *Column) Desc(d int) SqlExpr {
	return c.Expr("# - ?", d)
}

func (c *Column) Like(v string) SqlCondition {
	return AsCond(c.Expr("# LIKE ?", "%"+v+"%"))
}

func (c *Column) LeftLike(v string) SqlCondition {
	return AsCond(c.Expr("# LIKE ?", "%"+v))
}

func (c *Column) RightLike(v string) SqlCondition {
	return AsCond(c.Expr("# LIKE ?", v+"%"))
}

func (c *Column) NotLike(v string) SqlCondition {
	return AsCond(c.Expr("# NOT LIKE ?", "%"+v+"%"))
}

func (c *Column) IsNull() SqlCondition {
	return AsCond(c.Expr("# IS NULL"))
}

func (c *Column) IsNotNull() SqlCondition {
	return AsCond(c.Expr("# IS NOT NULL"))
}

func (c *Column) Between(leftValue interface{}, rightValue interface{}) SqlCondition {
	return AsCond(c.Expr("# BETWEEN ? AND ?", leftValue, rightValue))
}

func (c *Column) NotBetween(leftValue interface{}, rightValue interface{}) SqlCondition {
	return AsCond(c.Expr("# NOT BETWEEN ? AND ?", leftValue, rightValue))
}

func (c *Column) In(args ...interface{}) SqlCondition {
	length := len(args)
	if length == 0 {
		return nil
	}

	e := Expr("# IN ")
	e.WriteGroup(func(e *Ex) {
		for i := 0; i < length; i++ {
			e.WriteHolder(i)
		}
	})

	return AsCond(c.Expr(e.String(), args...))
}

func (c *Column) NotIn(args ...interface{}) SqlCondition {
	length := len(args)
	if length == 0 {
		return nil
	}

	e := Expr("# NOT IN ")
	e.WriteGroup(func(e *Ex) {
		for i := 0; i < length; i++ {
			e.WriteHolder(i)
		}
	})

	return AsCond(c.Expr(e.String(), args...))
}

func (c *Column) Eq(v interface{}) SqlCondition {
	return AsCond(c.Expr("# = ?", v))
}

func (c *Column) Neq(v interface{}) SqlCondition {
	return AsCond(c.Expr("# <> ?", v))
}

func (c *Column) Gt(v interface{}) SqlCondition {
	return AsCond(c.Expr("# > ?", v))
}

func (c *Column) Gte(v interface{}) SqlCondition {
	return AsCond(c.Expr("# >= ?", v))
}

func (c *Column) Lt(v interface{}) SqlCondition {
	return AsCond(c.Expr("# < ?", v))
}

func (c *Column) Lte(v interface{}) SqlCondition {
	return AsCond(c.Expr("# <= ?", v))
}

func ColumnTypeFromTypeAndTag(typ reflect.Type, nameAndFlags string) *ColumnType {
	ct := &ColumnType{}
	ct.Type = reflectx.Deref(typ)

	v := reflect.New(ct.Type).Interface()

	if dataTypeDescriber, ok := v.(DataTypeDescriber); ok {
		ct.GetDataType = dataTypeDescriber.DataType
	}

	if strings.Index(nameAndFlags, ",") > -1 {
		for _, flag := range strings.Split(nameAndFlags, ",")[1:] {
			nameAndValue := strings.Split(flag, "=")
			switch strings.ToLower(nameAndValue[0]) {
			case "null":
				ct.Null = true
			case "autoincrement":
				ct.AutoIncrement = true
			case "deprecated":
				rename := ""
				if len(nameAndValue) > 1 {
					rename = nameAndValue[1]
				}
				ct.DeprecatedActions = &DeprecatedActions{RenameTo: rename}
			case "size":
				if len(nameAndValue) == 1 {
					panic(fmt.Errorf("missing size value"))
				}
				length, err := strconv.ParseUint(nameAndValue[1], 10, 64)
				if err != nil {
					panic(fmt.Errorf("invalid size value: %s", err))
				}
				ct.Length = length
			case "decimal":
				if len(nameAndValue) == 1 {
					panic(fmt.Errorf("missing size value"))
				}
				decimal, err := strconv.ParseUint(nameAndValue[1], 10, 64)
				if err != nil {
					panic(fmt.Errorf("invalid decimal value: %s", err))
				}
				ct.Decimal = decimal
			case "default":
				if len(nameAndValue) == 1 {
					panic(fmt.Errorf("missing default value"))
				}
				ct.Default = &nameAndValue[1]
			}
		}
	}

	return ct
}

type ColumnType struct {
	Type        reflect.Type
	GetDataType func(engine string) string

	Length  uint64
	Decimal uint64

	Default *string

	Null          bool
	AutoIncrement bool

	Comment string

	DeprecatedActions *DeprecatedActions
}

type DeprecatedActions struct {
	RenameTo string `name:"rename"`
}
