package builder

import (
	"container/list"
	"fmt"
	"reflect"
	"strings"
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
	Exactly   bool

	Description []string
	Relation    []string

	*ColumnType
}

func (c *Column) Of(table *Table) *Column {
	col := &Column{
		Name:       c.Name,
		FieldName:  c.FieldName,
		Table:      table,
		Exactly:    true,
		ColumnType: c.ColumnType,
	}
	return col
}

func (c *Column) IsNil() bool {
	return c == nil
}

func (c *Column) Expr() *Ex {
	if c.Table != nil && c.Exactly {
		return Expr("?.?", c.Table, Expr(c.Name))
	}
	return Expr(c.Name)
}

func (c *Column) Ex(query string, args ...interface{}) *Ex {
	e := Expr("")

	qc := 0
	n := len(args)

	for _, key := range []byte(query) {
		switch key {
		case '#':
			e.WriteByte('?')
			e.AppendArgs(c.Expr())
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
	return AsAssignment(c.Ex("# = ?", v))
}

func (c *Column) Incr(d int) *Ex {
	return c.Ex("# + ?", d)
}

func (c *Column) Desc(d int) *Ex {
	return c.Ex("# - ?", d)
}

func (c *Column) Like(v string) *Condition {
	return AsCond(c.Ex("# LIKE ?", "%"+v+"%"))
}

func (c *Column) LeftLike(v string) *Condition {
	return AsCond(c.Ex("# LIKE ?", "%"+v))
}

func (c *Column) RightLike(v string) *Condition {
	return AsCond(c.Ex("# LIKE ?", v+"%"))
}

func (c *Column) NotLike(v string) *Condition {
	return AsCond(c.Ex("# NOT LIKE ?", "%"+v+"%"))
}

func (c *Column) IsNull() *Condition {
	return AsCond(c.Ex("# IS NULL"))
}

func (c *Column) IsNotNull() *Condition {
	return AsCond(c.Ex("# IS NOT NULL"))
}

func (c *Column) Between(leftValue interface{}, rightValue interface{}) *Condition {
	return AsCond(c.Ex("# BETWEEN ? AND ?", leftValue, rightValue))
}

func (c *Column) NotBetween(leftValue interface{}, rightValue interface{}) *Condition {
	return AsCond(c.Ex("# NOT BETWEEN ? AND ?", leftValue, rightValue))
}

func (c *Column) In(args ...interface{}) *Condition {
	length := len(args)
	if length == 0 {
		return nil
	}
	e := c.Expr()
	e.WriteString(" IN ")
	e.WriteGroup(func(e *Ex) {
		for i := range args {
			e.WriteHolder(i)
			e.AppendArgs(args[i])
		}
	})
	return AsCond(e)
}

func (c *Column) NotIn(args ...interface{}) *Condition {
	length := len(args)
	if length == 0 {
		return nil
	}
	e := c.Expr()
	e.WriteString(" NOT IN ")
	e.WriteGroup(func(e *Ex) {
		for i := range args {
			e.WriteHolder(i)
			e.AppendArgs(args[i])
		}
	})
	return AsCond(e)
}

func (c *Column) Eq(v interface{}) *Condition {
	return AsCond(c.Ex("# = ?", v))
}

func (c *Column) Neq(v interface{}) *Condition {
	return AsCond(c.Ex("# <> ?", v))
}

func (c *Column) Gt(v interface{}) *Condition {
	return AsCond(c.Ex("# > ?", v))
}

func (c *Column) Gte(v interface{}) *Condition {
	return AsCond(c.Ex("# >= ?", v))
}

func (c *Column) Lt(v interface{}) *Condition {
	return AsCond(c.Ex("# < ?", v))
}

func (c *Column) Lte(v interface{}) *Condition {
	return AsCond(c.Ex("# <= ?", v))
}

func Cols(names ...string) *Columns {
	cols := &Columns{}
	for _, name := range names {
		cols.Add(Col(name))
	}
	return cols
}

type Columns struct {
	l             *list.List
	columns       map[string]*list.Element
	fields        map[string]*list.Element
	autoIncrement *Column
}

func (cols *Columns) IsNil() bool {
	return cols == nil || cols.Len() == 0
}

func (cols *Columns) Expr() *Ex {
	expr := Expr("")

	cols.Range(func(col *Column, idx int) {
		if idx > 0 {
			expr.WriteByte(',')
		}
		expr.WriteExpr(col)
	})

	return expr
}

func (cols *Columns) AutoIncrement() (col *Column) {
	return cols.autoIncrement
}

func (cols *Columns) Clone() *Columns {
	c := &Columns{}
	cols.Range(func(col *Column, idx int) {
		c.Add(col)
	})
	return c
}

func (cols *Columns) Len() int {
	if cols == nil || cols.l == nil {
		return 0
	}
	return cols.l.Len()
}

func (cols *Columns) MustFields(fieldNames ...string) *Columns {
	nextCols, err := cols.Fields(fieldNames...)
	if err != nil {
		panic(err)
	}
	return nextCols
}

func (cols *Columns) Fields(fieldNames ...string) (*Columns, error) {
	if len(fieldNames) == 0 {
		return cols.Clone(), nil
	}
	newCols := &Columns{}
	for _, fieldName := range fieldNames {
		col := cols.F(fieldName)
		if col == nil {
			return nil, fmt.Errorf("unknown struct field %s", fieldName)
		}
		newCols.Add(col)
	}
	return newCols, nil
}

func (cols *Columns) FieldNames() []string {
	fieldNames := make([]string, 0)
	cols.Range(func(col *Column, idx int) {
		fieldNames = append(fieldNames, col.FieldName)
	})
	return fieldNames
}

func (cols *Columns) F(fileName string) (col *Column) {
	if cols.fields != nil {
		if c, ok := cols.fields[fileName]; ok {
			return c.Value.(*Column)
		}
	}
	return nil
}

func (cols *Columns) List() (l []*Column) {
	if cols != nil && cols.columns != nil {
		cols.Range(func(col *Column, idx int) {
			l = append(l, col)
		})
	}
	return
}

func (cols *Columns) Cols(colNames ...string) (*Columns, error) {
	if len(colNames) == 0 {
		return cols.Clone(), nil
	}
	newCols := &Columns{}
	for _, colName := range colNames {
		col := cols.Col(colName)
		if col == nil {
			return nil, fmt.Errorf("unknown struct column %s", colName)
		}
		newCols.Add(col)
	}
	return newCols, nil
}

func (cols *Columns) Col(columnName string) (col *Column) {
	columnName = strings.ToLower(columnName)
	if cols.columns != nil {
		if c, ok := cols.columns[columnName]; ok {
			return c.Value.(*Column)
		}
	}
	return nil
}

func (cols *Columns) Add(columns ...*Column) {
	if cols.columns == nil {
		cols.columns = map[string]*list.Element{}
		cols.fields = map[string]*list.Element{}
		cols.l = list.New()
	}

	for _, col := range columns {
		if col != nil {
			if col.ColumnType != nil && col.ColumnType.AutoIncrement {
				if cols.autoIncrement != nil {
					panic(fmt.Errorf("AutoIncrement field can only have one, now %s, but %s want to replace", cols.autoIncrement.Name, col.Name))
				}
				cols.autoIncrement = col
			}
			e := cols.l.PushBack(col)
			cols.columns[col.Name] = e
			cols.fields[col.FieldName] = e
		}
	}
}

func (cols *Columns) Remove(name string) {
	name = strings.ToLower(name)
	if cols.columns != nil {
		if e, exists := cols.columns[name]; exists {
			cols.l.Remove(e)
			delete(cols.columns, name)
		}
	}
}

func (cols *Columns) Range(cb func(col *Column, idx int)) {
	if cols.l != nil {
		i := 0
		for e := cols.l.Front(); e != nil; e = e.Next() {
			cb(e.Value.(*Column), i)
			i++
		}
	}
}
