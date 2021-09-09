package builder

import (
	"context"
	"fmt"
	"strings"
)

func Cols(names ...string) *Columns {
	cols := &Columns{}
	for _, name := range names {
		cols.Add(Col(name))
	}
	return cols
}

type Columns struct {
	l             []*Column
	autoIncrement *Column
}

func (cols *Columns) IsNil() bool {
	return cols == nil || cols.Len() == 0
}

func (cols *Columns) Ex(ctx context.Context) *Ex {
	e := Expr("")
	e.Grow(cols.Len())

	cols.Range(func(col *Column, idx int) {
		if idx > 0 {
			e.WriteQueryByte(',')
		}
		e.WriteExpr(col)
	})

	return e.Ex(ctx)
}

func (cols *Columns) AutoIncrement() (col *Column) {
	return cols.autoIncrement
}

func (cols *Columns) Clone() *Columns {
	c := &Columns{}

	n := len(cols.l)

	c.l = make([]*Column, n)

	for i := range c.l {
		c.l[i] = cols.l[i]
	}

	return c
}

func (cols *Columns) Len() int {
	if cols == nil || cols.l == nil {
		return 0
	}
	return len(cols.l)
}

func (cols *Columns) MustFields(structFieldNames ...string) *Columns {
	nextCols, err := cols.Fields(structFieldNames...)
	if err != nil {
		panic(err)
	}
	return nextCols
}

func (cols *Columns) Fields(structFieldNames ...string) (*Columns, error) {
	if len(structFieldNames) == 0 {
		return cols.Clone(), nil
	}

	c := &Columns{}
	c.l = make([]*Column, len(structFieldNames))

	for i, fieldName := range structFieldNames {
		col := cols.F(fieldName)
		if col == nil {
			return nil, fmt.Errorf("unknown struct field %s", fieldName)
		}
		c.l[i] = col
	}
	return c, nil
}

func (cols *Columns) FieldNames() []string {
	fieldNames := make([]string, 0, len(cols.l))
	cols.Range(func(col *Column, idx int) {
		if col.FieldName != "" {
			fieldNames = append(fieldNames, col.FieldName)
		}
	})
	return fieldNames
}

func (cols *Columns) ColNames() []string {
	colNames := make([]string, 0, len(cols.l))
	cols.Range(func(col *Column, idx int) {
		if col.Name != "" {
			colNames = append(colNames, col.Name)
		}
	})
	return colNames
}

func (cols *Columns) F(structFieldName string) (col *Column) {
	for i := range cols.l {
		e := cols.l[i]

		if structFieldName == e.FieldName {
			return e
		}
	}
	return nil
}

func (cols *Columns) List() (l []*Column) {
	if cols != nil && cols.l != nil {
		return cols.l
	}
	return nil
}

func (cols *Columns) MustCols(colNames ...string) *Columns {
	nextCols, err := cols.Cols(colNames...)
	if err != nil {
		panic(err)
	}
	return nextCols
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
	for i := range cols.l {
		c := cols.l[i]
		if columnName == c.Name {
			return c
		}
	}
	return nil
}

func (cols *Columns) Add(columns ...*Column) {
	for i := range columns {
		col := columns[i]
		if col == nil {
			continue
		}
		if col.ColumnType != nil && col.ColumnType.AutoIncrement {
			if cols.autoIncrement != nil {
				panic(fmt.Errorf("AutoIncrement field can only have one, now %s, but %s want to replace", cols.autoIncrement.Name, col.Name))
			}
			cols.autoIncrement = col
		}
		cols.l = append(cols.l, col)
	}
}

func (cols *Columns) Range(cb func(col *Column, idx int)) {
	for i := range cols.l {
		cb(cols.l[i], i)
	}
}
