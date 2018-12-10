package sqlx

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"reflect"

	"github.com/go-courier/sqlx/v2/builder"
)

type SchemaDescriber interface {
	Schema() string
}

func NewFeatureDatabase(name string) *Database {
	if projectFeature, exists := os.LookupEnv("PROJECT_FEATURE"); exists && projectFeature != "" {
		name = name + "__" + projectFeature
	}
	return NewDatabase(name)
}

func NewDatabase(name string) *Database {
	return &Database{
		Name:   name,
		Tables: builder.Tables{},
	}
}

type Database struct {
	Name   string
	Tables builder.Tables
}

type DBNameBinder interface {
	WithDBName(dbName string) driver.Connector
}

func (database *Database) OpenDB(connector driver.Connector) *DB {
	if dbNameBinder, ok := connector.(DBNameBinder); ok {
		connector = dbNameBinder.WithDBName(database.Name)
	}
	dialet, ok := connector.(builder.Dialect)
	if !ok {
		panic(fmt.Errorf("connector should implement builder.Dialect"))
	}
	return &DB{
		Dialect:     dialet,
		SqlExecutor: sql.OpenDB(connector),
	}
}

func (database *Database) AddTable(table *builder.Table) {
	database.Tables.Add(table)
}

func (database *Database) Register(model builder.Model) *builder.Table {
	tpe := reflect.TypeOf(model)
	if tpe.Kind() != reflect.Ptr {
		panic(fmt.Errorf("model %s must be a pointer", tpe.Name()))
	}
	tpe = tpe.Elem()
	if tpe.Kind() != reflect.Struct {
		panic(fmt.Errorf("model %s must be a struct", tpe.Name()))
	}
	table := builder.T(model.TableName())
	builder.ScanDefToTable(reflect.Indirect(reflect.ValueOf(model)), table)
	database.AddTable(table)
	return table
}

func (database *Database) Table(tableName string) *builder.Table {
	if t, ok := database.Tables[tableName]; ok {
		return t
	}
	return nil
}

func (database *Database) T(model builder.Model) *builder.Table {
	t := database.Table(model.TableName())

	if sd, ok := model.(SchemaDescriber); ok {
		schema := sd.Schema()
		if schema != "" {
			return t.WithSchema(schema)
		}
	}

	return t
}

func (database *Database) Assignments(model builder.Model, zeroFields ...string) builder.Assignments {
	table := database.T(model)
	return table.AssignmentsByFieldValues(database.FieldValuesFromModel(table, model, zeroFields...))
}

func (database *Database) Insert(model builder.Model, zeroFields []string, additions ...builder.Addition) builder.SqlExpr {
	table := database.T(model)
	cols, vals := table.ColumnsAndValuesByFieldValues(database.FieldValuesFromModel(table, model, zeroFields...))
	return builder.Insert().Into(table, additions...).Values(cols, vals...)
}

func (database *Database) FieldValuesFromModel(table *builder.Table, model builder.Model, zeroFields ...string) builder.FieldValues {
	fieldValues := builder.FieldValuesFromStructByNonZero(model, zeroFields...)
	if autoIncrementCol := table.AutoIncrement(); autoIncrementCol != nil {
		delete(fieldValues, autoIncrementCol.FieldName)
	}
	return fieldValues
}
