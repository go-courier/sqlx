package sqlx

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"

	"github.com/go-courier/sqlx/v2/builder"
)

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
	Schema string
	Tables builder.Tables
}

func (database Database) WithSchema(schema string) *Database {
	database.Schema = schema

	tables := builder.Tables{}

	database.Tables.Range(func(tab *builder.Table, idx int) {
		tables.Add(tab.WithSchema(database.Schema))
	})

	database.Tables = tables

	return &database
}

type DBNameBinder interface {
	WithDBName(dbName string) driver.Connector
}

func (database *Database) OpenDB(connector driver.Connector) *DB {
	if dbNameBinder, ok := connector.(DBNameBinder); ok {
		connector = dbNameBinder.WithDBName(database.Name)
	}
	dialect, ok := connector.(builder.Dialect)
	if !ok {
		panic(fmt.Errorf("connector should implement builder.Dialect"))
	}
	return &DB{
		Database:    database,
		dialect:     dialect,
		SqlExecutor: sql.OpenDB(connector),
	}
}

func (database *Database) AddTable(table *builder.Table) {
	database.Tables.Add(table)
}

func (database *Database) Register(model builder.Model) *builder.Table {
	table := builder.TableFromModel(model)
	table.Schema = database.Schema
	database.AddTable(table)
	return table
}

func (database *Database) Table(tableName string) *builder.Table {
	return database.Tables.Table(tableName)
}

func (database *Database) T(model builder.Model) *builder.Table {
	if td, ok := model.(builder.TableDefinition); ok {
		return td.T()
	}

	if t, ok := model.(*builder.Table); ok {
		return t
	}

	return database.Table(model.TableName())
}
