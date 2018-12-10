package builder

import (
	"regexp"
	"sort"
	"strings"
)

type TableDefinition interface {
	T() *Table
}

func T(tableName string, tableDefinitions ...TableDefinition) *Table {
	t := &Table{
		Name: tableName,
	}

	for _, tableDef := range tableDefinitions {
		switch d := tableDef.(type) {
		case *Column:
			t.AddCol(d)
		}
	}
	for _, tableDef := range tableDefinitions {
		switch d := tableDef.(type) {
		case *Key:
			t.AddKey(d)
		}
	}
	return t
}

type Table struct {
	Name   string
	Schema string
	Model  Model
	Columns
	Keys
}

func (t *Table) IsNil() bool {
	return t == nil || len(t.Name) == 0
}

func (t Table) WithSchema(schema string) *Table {
	t.Schema = schema
	return &t
}

func (t *Table) Expr() *Ex {
	if t.Schema != "" {
		return Expr(t.Schema + "." + t.Name)
	}
	return Expr(t.Name)
}

func (t *Table) AddCol(d *Column) {
	if d == nil {
		return
	}
	t.Columns.Add(d.On(t))
}

func (t *Table) AddKey(key *Key) {
	if key == nil {
		return
	}
	t.Keys.Add(key.On(t))
}

var fieldNamePlaceholder = regexp.MustCompile("#[A-Z][A-Za-z0-9_]+")

// replace go struct field name with table column name
func (t *Table) Ex(query string, args ...interface{}) *Ex {
	finalQuery := fieldNamePlaceholder.ReplaceAllStringFunc(query, func(i string) string {
		fieldName := strings.TrimLeft(i, "#")
		if col := t.F(fieldName); col != nil {
			return col.Name
		}
		return i
	})
	return Expr(finalQuery, args...)
}

func (t *Table) ColumnsAndValuesByFieldValues(fieldValues FieldValues) (columns *Columns, args []interface{}) {
	fieldNames := make([]string, 0)
	for fieldName := range fieldValues {
		fieldNames = append(fieldNames, fieldName)
	}

	sort.Strings(fieldNames)

	columns = &Columns{}

	for _, fieldName := range fieldNames {
		if col := t.F(fieldName); col != nil {
			columns.Add(col)
			args = append(args, fieldValues[fieldName])
		}
	}
	return
}

func (t *Table) AssignmentsByFieldValues(fieldValues FieldValues) (assignments Assignments) {
	for fieldName, value := range fieldValues {
		col := t.F(fieldName)
		if col != nil {
			assignments = append(assignments, col.ValueBy(value))
		}
	}
	return
}

func (t *Table) Diff(prevTable *Table, dialect Dialect, skipDropColumn bool) (exprList []SqlExpr) {
	cols := map[string]bool{}

	// add or modify columns
	t.Columns.Range(func(col *Column, idx int) {
		cols[col.Name] = true
		if prevTable.Col(col.Name) == nil {
			exprList = append(exprList, dialect.AddColumn(col))
		} else {
			exprList = append(exprList, dialect.ModifyColumn(col))
		}
	})

	{
		indexes := map[string]bool{}

		t.Keys.Range(func(key *Key, idx int) {
			name := key.Name
			if key.IsPrimary() {
				name = dialect.PrimaryKeyName()
			}
			indexes[name] = true

			prevKey := prevTable.Key(name)
			if prevKey == nil {
				exprList = append(exprList, dialect.AddIndex(key))
			} else {
				if !key.IsPrimary() && key.Columns.Expr().Query() != prevTable.Columns.Expr().Query() {
					exprList = append(exprList, dialect.DropIndex(key))
					exprList = append(exprList, dialect.AddIndex(key))
				}
			}
		})

		prevTable.Keys.Range(func(key *Key, idx int) {
			if _, ok := indexes[strings.ToLower(key.Name)]; !ok {
				exprList = append(exprList, dialect.DropIndex(key))
			}
		})
	}

	// drop columns
	if !skipDropColumn {
		prevTable.Columns.Range(func(col *Column, idx int) {
			if _, ok := cols[strings.ToLower(col.Name)]; !ok {
				exprList = append(exprList, dialect.DropColumn(col))
			}
		})
	}

	return
}

type Tables map[string]*Table

func (tables Tables) TableNames() (names []string) {
	for name := range tables {
		names = append(names, name)
	}
	return
}

func (tables Tables) Add(table *Table) {
	tables[table.Name] = table
}
