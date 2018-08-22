package builder

import (
	"fmt"
)

type DatabaseDefiner interface {
	DBName() string
}

func CreateDatebase(d DatabaseDefiner) SqlExpr {
	return Expr(fmt.Sprintf("CREATE DATABASE %s", quote(d.DBName())))
}

func CreateDatebaseIfNotExists(d DatabaseDefiner) SqlExpr {
	return Expr(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", quote(d.DBName())))
}

func DropDatabase(d DatabaseDefiner) SqlExpr {
	return Expr(fmt.Sprintf("DROP DATABASE %s", quote(d.DBName())))
}

func DB(name string) *Database {
	return &Database{
		Name:   name,
		Tables: Tables{},
	}
}

type Database struct {
	Name string
	Tables
}

func (d *Database) DBName() string {
	return d.Name
}

func (d *Database) Register(table *Table) *Database {
	d.Tables.Add(table)
	table.DB = d
	return d
}

func (d *Database) Table(name string) *Table {
	if table, ok := d.Tables[name]; ok {
		return table
	}
	return nil
}
