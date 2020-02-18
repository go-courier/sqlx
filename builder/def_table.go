package builder

import (
	"bytes"
	"container/list"
	"context"
	"fmt"
	"sort"
	"strings"
	"text/scanner"
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
	Name        string
	Description []string

	Schema    string
	ModelName string
	Model     Model

	Columns
	Keys
}

func (t *Table) TableName() string {
	return t.Name
}

func (t *Table) IsNil() bool {
	return t == nil || len(t.Name) == 0
}

func (t Table) WithSchema(schema string) *Table {
	t.Schema = schema

	cols := Columns{}
	t.Columns.Range(func(col *Column, idx int) {
		cols.Add(col.On(&t))
	})
	t.Columns = cols

	keys := Keys{}
	t.Keys.Range(func(key *Key, idx int) {
		keys.Add(key.On(&t))
	})
	t.Keys = keys

	return &t
}

func (t *Table) Ex(ctx context.Context) *Ex {
	if t.Schema != "" {
		return Expr(t.Schema + "." + t.Name).Ex(ctx)
	}
	return Expr(t.Name).Ex(ctx)
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

func (t *Table) Expr(query string, args ...interface{}) *Ex {
	if query == "" {
		return nil
	}

	e := Expr("")

	s := &scanner.Scanner{}
	s.Init(bytes.NewBuffer([]byte(query)))

	fieldNameBuf := bytes.NewBuffer(nil)

	n := len(args)
	queryCount := 0

	for tok := s.Next(); tok != scanner.EOF; tok = s.Next() {
		switch tok {
		case '#':
			fieldNameBuf = bytes.NewBuffer(nil)

			e.WriteHolder(0)

			for {
				tok = s.Next()

				if tok == scanner.EOF {
					break
				}

				if (tok >= 'A' && tok <= 'Z') ||
					(tok >= 'a' && tok <= 'z') ||
					(tok >= '0' && tok <= '9') ||
					tok == '_' {

					fieldNameBuf.WriteRune(tok)
					continue
				}

				e.WriteRune(tok)

				break
			}

			//spew.Dump(fieldNameBuf.String())

			if fieldNameBuf.Len() == 0 {
				e.AppendArgs(t)
			} else {
				fieldName := fieldNameBuf.String()
				col := t.F(fieldNameBuf.String())
				if col == nil {
					panic(fmt.Errorf("missing field fieldName %s of table %s", fieldName, t.Name))
				}
				e.AppendArgs(col)
			}
		case '?':
			e.WriteRune(tok)
			if queryCount < n {
				e.AppendArgs(args[queryCount])
				queryCount++
			}
		default:
			e.WriteRune(tok)
		}
	}

	return e
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

func (t *Table) Diff(prevTable *Table, dialect Dialect) (exprList []SqlExpr) {
	// diff columns
	t.Columns.Range(func(col *Column, idx int) {
		if prevTable.Col(col.Name) != nil {
			currentCol := t.Col(col.Name)
			if currentCol != nil {
				if currentCol.DeprecatedActions != nil {
					renameTo := currentCol.DeprecatedActions.RenameTo
					if renameTo != "" {
						prevCol := prevTable.Col(renameTo)
						if prevCol != nil {
							exprList = append(exprList, dialect.DropColumn(prevCol))
						}
						targetCol := t.Col(renameTo)
						if targetCol == nil {
							panic(fmt.Errorf("col `%s` is not declared", renameTo))
						}

						exprList = append(exprList, dialect.RenameColumn(col, targetCol))
						prevTable.AddCol(targetCol)
						return
					}
					exprList = append(exprList, dialect.DropColumn(col))
					return
				}
				if ResolveExpr(col).Query() != ResolveExpr(currentCol).Query() {
					exprList = append(exprList, dialect.ModifyColumn(col))
				}
				return
			}
			exprList = append(exprList, dialect.DropColumn(col))
			return
		}

		if col.DeprecatedActions == nil {
			exprList = append(exprList, dialect.AddColumn(col))
		}
	})

	// indexes
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
			if !key.IsPrimary() && ResolveExpr(key.Columns).Query() != ResolveExpr(prevKey.Columns).Query() {
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

	return
}

type Tables struct {
	l      *list.List
	tables map[string]*list.Element
	models map[string]*list.Element
}

func (tables *Tables) TableNames() (names []string) {
	tables.Range(func(tab *Table, idx int) {
		names = append(names, tab.Name)
	})
	return
}

func (tables *Tables) Add(tabs ...*Table) {
	if tables.tables == nil {
		tables.tables = map[string]*list.Element{}
		tables.models = map[string]*list.Element{}
		tables.l = list.New()
	}

	for _, tab := range tabs {
		if tab != nil {
			if _, ok := tables.tables[tab.Name]; ok {
				tables.Remove(tab.Name)
			}

			e := tables.l.PushBack(tab)
			tables.tables[tab.Name] = e
			if tab.ModelName != "" {
				tables.models[tab.ModelName] = e
			}
		}
	}
}

func (tables *Tables) Table(tableName string) *Table {
	if tables.tables != nil {
		if c, ok := tables.tables[tableName]; ok {
			return c.Value.(*Table)
		}
	}
	return nil
}

func (tables *Tables) Model(structName string) *Table {
	if tables.models != nil {
		if c, ok := tables.models[structName]; ok {
			return c.Value.(*Table)
		}
	}
	return nil
}

func (tables *Tables) Remove(name string) {
	if tables.tables != nil {
		if e, exists := tables.tables[name]; exists {
			tables.l.Remove(e)
			delete(tables.tables, name)
		}
	}
}

func (tables *Tables) Range(cb func(tab *Table, idx int)) {
	if tables.l != nil {
		i := 0
		for e := tables.l.Front(); e != nil; e = e.Next() {
			cb(e.Value.(*Table), i)
			i++
		}
	}
}
