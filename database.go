package sqlx

import (
	"fmt"
	"os"
	"reflect"

	"github.com/go-courier/sqlx/builder"
)

func NewFeatureDatabase(name string) *Database {
	if projectFeature, exists := os.LookupEnv("PROJECT_FEATURE"); exists && projectFeature != "" {
		name = name + "__" + projectFeature
	}
	return NewDatabase(name)
}

func NewDatabase(name string) *Database {
	return &Database{
		Database: builder.DB(name),
	}
}

type Database struct {
	*builder.Database
}

func (database *Database) Register(model builder.Model) *builder.Table {
	database.mustStructType(model)

	table := builder.T(database.Database, model.TableName())
	builder.ScanDefToTable(reflect.Indirect(reflect.ValueOf(model)), table)

	database.Database.Register(table)
	return table
}

func (database Database) T(model builder.Model) *builder.Table {
	database.mustStructType(model)
	return database.Database.Table(model.TableName())
}

func (Database) mustStructType(model builder.Model) {
	tpe := reflect.TypeOf(model)
	if tpe.Kind() != reflect.Ptr {
		panic(fmt.Errorf("model %s must be a pointer", tpe.Name()))
	}
	tpe = tpe.Elem()
	if tpe.Kind() != reflect.Struct {
		panic(fmt.Errorf("model %s must be a struct", tpe.Name()))

	}
}

func (database *Database) Assignments(model builder.Model, excludes ...string) builder.Assignments {
	table := database.T(model)
	fieldValues := builder.FieldValuesFromStructByNonZero(model, excludes...)
	if autoIncrementCol := table.AutoIncrement(); autoIncrementCol != nil {
		delete(fieldValues, autoIncrementCol.FieldName)
	}
	return table.AssignmentsByFieldValues(fieldValues)
}
