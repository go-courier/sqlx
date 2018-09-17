package builder

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type TableDef interface {
	IsValidDef() bool
	Def() *Expression
}

func CreateTable(t *Table) SqlExpr {
	return createTable(t, false)
}

func CreateTableIsNotExists(t *Table) SqlExpr {
	return createTable(t, true)
}

func createTable(t *Table, ifNotExists bool) SqlExpr {
	expr := Expr("CREATE TABLE")
	if ifNotExists {
		expr = MustJoinExpr(" ", expr, Expr("IF NOT EXISTS"))
	}
	expr.Query = expr.Query + fmt.Sprintf(" %s (", t.FullName())

	if !t.Columns.IsEmpty() {
		isFirstCol := true

		t.Columns.Range(func(col *Column, idx int) {
			joiner := ", "
			if isFirstCol {
				joiner = ""
			}
			def := col.Def()
			if def != nil {
				isFirstCol = false
				expr = MustJoinExpr(joiner, expr, col.Def())
			}
		})

		t.Keys.Range(func(key *Key, idx int) {
			expr = MustJoinExpr(", ", expr, key.Def())
		})
	}

	engine := t.Engine
	if engine == "" {
		engine = "InnoDB"
	}

	charset := t.Charset
	if charset == "" {
		charset = "utf8"
	}

	expr.Query = fmt.Sprintf("%s) ENGINE=%s CHARSET=%s", expr.Query, engine, charset)
	return expr
}

func AlterTable(t *Table) SqlExpr {
	return Expr(fmt.Sprintf("ALTER TABLE %s", t.FullName()))
}

func DropTable(t *Table) SqlExpr {
	return Expr(fmt.Sprintf("DROP TABLE %s", t.FullName()))
}

func TruncateTable(t *Table) SqlExpr {
	return Expr(fmt.Sprintf("TRUNCATE TABLE %s", t.FullName()))
}

func T(db *Database, tableName string) *Table {
	return &Table{
		DB:   db,
		Name: tableName,
	}
}

type Table struct {
	DB   *Database
	Name string
	Columns
	Keys
	Engine  string
	Charset string
}

func (t *Table) Expr() *Expression {
	return Expr(t.FullName())
}

func (t Table) Define(defs ...TableDef) *Table {
	for _, def := range defs {
		if def.IsValidDef() {
			switch def.(type) {
			case *Column:
				t.Columns.Add(def.(*Column))
			case *Key:
				t.Keys.Add(def.(*Key))
			}
		}
	}
	return &t
}

var (
	fieldNamePlaceholder = regexp.MustCompile("#[A-Z][A-Za-z0-9_]+")
)

// replace go struct field name with table column name
func (t *Table) Ex(query string, args ...interface{}) *Expression {
	finalQuery := fieldNamePlaceholder.ReplaceAllStringFunc(query, func(i string) string {
		fieldName := strings.TrimLeft(i, "#")
		if col := t.F(fieldName); col != nil {
			return col.String()
		}
		return i
	})
	return Expr(finalQuery, args...)
}

func (t *Table) Cond(query string, args ...interface{}) *Condition {
	return (*Condition)(t.Ex(query, args...))
}

type FieldValues map[string]interface{}

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

func (t *Table) TableName() string {
	return t.Name
}

func (t *Table) FullName() string {
	return quote(t.DB.DBName()) + "." + quote(t.TableName())
}

type DiffOptions struct {
	DropColumn bool
}

func (t *Table) Diff(table *Table, opts DiffOptions) SqlExpr {
	colsDiffResult := t.Columns.Diff(table.Columns)
	keysDiffResult := t.Keys.Diff(table.Keys)

	colsChanged := colsDiffResult.IsChanged()
	indexesChanged := keysDiffResult.IsChanged()

	if !colsChanged && !indexesChanged {
		return nil
	}
	expr := AlterTable(t)

	joiner := " "

	if colsChanged {
		if opts.DropColumn {
			colsDiffResult.colsForDelete.Range(func(col *Column, idx int) {
				expr = MustJoinExpr(joiner, expr, DropCol(col))
				joiner = ", "
			})
		}
		colsDiffResult.colsForUpdate.Range(func(col *Column, idx int) {
			expr = MustJoinExpr(joiner, expr, ModifyCol(col))
			joiner = ", "
		})
		colsDiffResult.colsForAdd.Range(func(col *Column, idx int) {
			expr = MustJoinExpr(joiner, expr, AddCol(col))
			joiner = ", "
		})
	}

	if indexesChanged {
		keysDiffResult.keysForDelete.Range(func(key *Key, idx int) {
			expr = MustJoinExpr(joiner, expr, DropKey(key))
			joiner = ", "
		})
		keysDiffResult.keysForUpdate.Range(func(key *Key, idx int) {
			expr = MustJoinExpr(joiner, expr, DropKey(key))
			joiner = ", "
			expr = MustJoinExpr(joiner, expr, AddKey(key))
		})
		keysDiffResult.keysForAdd.Range(func(key *Key, idx int) {
			expr = MustJoinExpr(joiner, expr, AddKey(key))
			joiner = ", "
		})
	}

	return expr
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
