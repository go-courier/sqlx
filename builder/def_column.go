package builder

import (
	"container/list"
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/go-courier/enumeration"
	"github.com/go-courier/sqlx/datatypes"
)

func AddCol(c *Column) *Expression {
	return MustJoinExpr(" ", Expr("ADD COLUMN"), c.Def())
}

func DropCol(c *Column) *Expression {
	return Expr(fmt.Sprintf("DROP COLUMN %s", c.String()))
}

func ModifyCol(c *Column) *Expression {
	return MustJoinExpr(" ", Expr("MODIFY COLUMN"), c.Def())
}

func Col(table *Table, columnName string) *Column {
	return &Column{
		Table: table,
		Name:  columnName,
	}
}

func Cols(t *Table, columnNames ...string) (*Columns) {
	cols := &Columns{}
	for _, columnName := range columnNames {
		cols.Add(Col(t, columnName))
	}
	return cols
}

type Columns struct {
	columns       map[string]*list.Element
	fields        map[string]*list.Element
	l             *list.List
	autoIncrement *Column
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
	if cols.l == nil {
		return 0
	}
	return cols.l.Len()
}

func (cols *Columns) IsEmpty() bool {
	return cols.l == nil || cols.l.Len() == 0
}

func (cols *Columns) Fields(fieldNames ...string) (columns *Columns) {
	if len(fieldNames) == 0 {
		return cols.Clone()
	}
	columns = &Columns{}
	for _, fieldName := range fieldNames {
		if col := cols.F(fieldName); col != nil {
			columns.Add(col)
		}
	}
	return
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
	if cols.columns != nil {
		cols.Range(func(col *Column, idx int) {
			l = append(l, col)
		})
	}
	return
}

func (cols *Columns) Cols(colNames ...string) (columns *Columns) {
	if len(colNames) == 0 {
		return cols.Clone()
	}
	columns = &Columns{}
	for _, colName := range colNames {
		if col := cols.Col(colName); col != nil {
			columns.Add(col)
		}
	}
	return
}

func (cols *Columns) Col(columnName string) (col *Column) {
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
			if col.AutoIncrement {
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

func (cols Columns) Group() (e *Expression) {
	e = cols.Expr()
	if e.Query != "" {
		e.Query = "(" + e.Query + ")"
	}
	return e
}

func (cols Columns) Expr() (e *Expression) {
	query := ""

	cols.Range(func(col *Column, idx int) {
		if idx == 0 {
			query = query + col.String()
		} else {
			query = query + "," + col.String()
		}
	})

	return Expr(query)
}

func (cols Columns) Diff(targetCols Columns) columnsDiffResult {
	r := columnsDiffResult{}

	cs := cols.Clone()

	targetCols.Range(func(col *Column, idx int) {
		if currentCol := cs.Col(col.Name); currentCol != nil {
			sqlCurrent := currentCol.ColumnType.DeAlias().String()
			sqlNext := col.ColumnType.DeAlias().String()
			if sqlCurrent != sqlNext {
				logrus.Warnf("data type of %s.%s current:`%s`, will be `%s`", col.Table.Name, col.Name, sqlCurrent, sqlNext)
				r.colsForUpdate.Add(col)
			}
		} else {
			r.colsForAdd.Add(col)
		}
		cs.Remove(col.Name)
	})

	cs.Range(func(col *Column, idx int) {
		r.colsForDelete.Add(col)
	})

	return r
}

type columnsDiffResult struct {
	colsForAdd    Columns
	colsForUpdate Columns
	colsForDelete Columns
}

func (r columnsDiffResult) IsChanged() bool {
	return !r.colsForAdd.IsEmpty() || !r.colsForUpdate.IsEmpty() || !r.colsForDelete.IsEmpty()
}

var _ TableDef = (*Column)(nil)

type Column struct {
	Table     *Table
	Name      string
	FieldName string
	datatypes.ColumnType
}

func (c Column) Field(fieldName string) *Column {
	c.FieldName = fieldName
	return &c
}

func (c Column) Type(tpe string) *Column {
	columnType, err := datatypes.ParseColumnType(tpe)
	if err != nil {
		panic(fmt.Errorf("%s %s", c.Name, err))
	}
	c.ColumnType = *columnType
	return &c
}

func (c Column) Enum(enum enumeration.Enum) *Column {
	c.ColumnType.Enum = enum
	return &c
}

func (c *Column) IsValidDef() bool {
	return c.ColumnType.DataType != ""
}

func (c *Column) Def() *Expression {
	return Expr(c.String() + " " + c.ColumnType.String())
}

func (c *Column) String() string {
	return quote(c.Name)
}

func (c *Column) Expr() *Expression {
	return Expr(c.String())
}

func (c *Column) ValueBy(v interface{}) *Assignment {
	if e, ok := v.(*Expression); ok {
		return (*Assignment)(Expr(fmt.Sprintf("%s = %s", c, e.Query), e.Args...))
	}
	return (*Assignment)(Expr(fmt.Sprintf("%s = ?", c), v))
}

func (c *Column) Incr(d int) *Expression {
	return Expr(fmt.Sprintf("%s + ?", c), d)
}

func (c *Column) Desc(d int) *Expression {
	return Expr(fmt.Sprintf("%s - ?", c), d)
}

func (c *Column) Like(v string) *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s LIKE ?", c), "%"+v+"%"))
}

func (c *Column) LeftLike(v string) *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s LIKE ?", c), "%"+v))
}

func (c *Column) RightLike(v string) *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s LIKE ?", c), v+"%"))
}

func (c *Column) NotLike(v string) *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s NOT LIKE ?", c), "%"+v+"%"))
}

func (c *Column) IsNull() *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s IS NULL", c)))
}

func (c *Column) IsNotNull() *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s IS NOT NULL", c)))
}

func (c *Column) Between(leftValue interface{}, rightValue interface{}) *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s BETWEEN ? AND ?", c), leftValue, rightValue))
}

func (c *Column) NotBetween(leftValue interface{}, rightValue interface{}) *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s NOT BETWEEN ? AND ?", c), leftValue, rightValue))
}

func (c *Column) In(args ...interface{}) *Condition {
	length := len(args)
	if length == 0 {
		return nil
	}
	if length == 1 {
		expr := ExprFrom(args[0])
		if expr != nil {
			return (*Condition)(Expr(fmt.Sprintf("%s IN (%s)", c, expr.Query), expr.Args...))
		}
	}
	return (*Condition)(Expr(fmt.Sprintf("%s IN (%s)", c, HolderRepeat(length)), args...))
}

func (c *Column) NotIn(args ...interface{}) *Condition {
	length := len(args)
	if length == 0 {
		return nil
	}
	if length == 1 {
		expr := ExprFrom(args[0])
		if expr != nil {
			return (*Condition)(Expr(fmt.Sprintf("%s NOT IN (%s)", c, expr.Query), expr.Args...))
		}
	}
	return (*Condition)(Expr(fmt.Sprintf("%s NOT IN (%s)", c, HolderRepeat(length)), args...))
}

func (c *Column) Eq(value interface{}) *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s = ?", c), value))
}

func (c *Column) Neq(v interface{}) *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s <> ?", c), v))
}

func (c *Column) Gt(v interface{}) *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s > ?", c), v))
}

func (c *Column) Gte(v interface{}) *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s >= ?", c), v))
}

func (c *Column) Lt(v interface{}) *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s < ?", c), v))
}

func (c *Column) Lte(v interface{}) *Condition {
	return (*Condition)(Expr(fmt.Sprintf("%s <= ?", c), v))
}
