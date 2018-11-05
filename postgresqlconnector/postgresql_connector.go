package postgresqlconnector

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/sqlx/v2/migration"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/go-courier/sqlx/v2"
)

var _ interface {
	driver.Connector
	builder.Dialect
} = (*PostgreSQLConnector)(nil)

type PostgreSQLConnector struct {
	Host       string
	DBName     string
	Extra      string
	Extensions []string
}

func dsn(host string, dbName string, extra string) string {
	if extra != "" {
		extra = "?" + extra
	}
	return host + "/" + dbName + extra
}

func (c PostgreSQLConnector) WithDBName(dbName string) driver.Connector {
	c.DBName = dbName
	return &c
}

func (c *PostgreSQLConnector) Migrate(db *sqlx.DB, database *sqlx.Database, opts *migration.MigrationOpts) error {
	if opts == nil {
		opts = &migration.MigrationOpts{}
	}

	prevDB := DBFromInformationSchema(db, database.Name, database.Tables.TableNames()...)

	if prevDB == nil {
		prevDB = &sqlx.Database{
			Name: database.Name,
		}
		if _, err := db.ExecExpr(db.CreateDatabase(database.Name)); err != nil {
			return err
		}
	}

	for name, table := range database.Tables {
		prevTable := prevDB.Table(name)
		if prevTable == nil {
			for _, expr := range db.CreateTableIsNotExists(table) {
				if _, err := db.ExecExpr(expr); err != nil {
					return err
				}
			}
			continue
		}

		exprList := table.Diff(prevTable, db.Dialect, opts.SkipDropColumn)

		for _, expr := range exprList {
			if !(expr == nil || expr.IsNil()) {
				if opts.DryRun {
					log.Printf(builder.ExprFrom(expr).Flatten().Query())
				} else {
					if _, err := db.ExecExpr(expr); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (c *PostgreSQLConnector) Connect(ctx context.Context) (driver.Conn, error) {
	d := c.Driver()
	conn, err := d.Open(dsn(c.Host, c.DBName, c.Extra))
	if err != nil {
		if c.IsErrorUnknownDatabase(err) {
			conn, err := d.Open(dsn(c.Host, "", c.Extra))
			if err != nil {
				return nil, err
			}
			stmt, _ := conn.Prepare(c.CreateDatabase(c.DBName).Expr().Query())
			if _, err := stmt.Exec(nil); err != nil {
				return nil, err
			}
			if err := stmt.Close(); err != nil {
				return nil, err
			}
			if err := conn.Close(); err != nil {
				return nil, err
			}
			return c.Connect(ctx)
		}
		return nil, err
	}
	for _, ex := range c.Extensions {
		stmt, _ := conn.Prepare("CREATE EXTENSION IF NOT EXISTS " + ex + ";")
		if _, err := stmt.Exec(nil); err != nil {
			return nil, err
		}
		if err := stmt.Close(); err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func (PostgreSQLConnector) Driver() driver.Driver {
	return &PostgreSQLLoggingDriver{Driver: &pq.Driver{}, Logger: logrus.StandardLogger()}
}

func (PostgreSQLConnector) BindVar(i int) string {
	return "$" + strconv.FormatInt(int64(i+1), 10)
}

func (PostgreSQLConnector) DriverName() string {
	return "postgres"
}

func (PostgreSQLConnector) PrimaryKeyName() string {
	return "pkey"
}

func (PostgreSQLConnector) IsErrorUnknownDatabase(err error) bool {
	if e, ok := err.(*pq.Error); ok && e.Code == "3D000" {
		return true
	}
	return false
}

func (PostgreSQLConnector) IsErrorConflict(err error) bool {
	if e, ok := err.(*pq.Error); ok && e.Code == "23505" {
		return true
	}
	return false
}

func (c *PostgreSQLConnector) CreateDatabase(dbName string) builder.SqlExpr {
	e := builder.Expr("CREATE DATABASE ")
	e.WriteString(dbName)
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) DropDatabase(dbName string) builder.SqlExpr {
	e := builder.Expr("DROP DATABASE ")
	e.WriteString(dbName)
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) AddIndex(key *builder.Key) builder.SqlExpr {
	if key.IsPrimary() {
		e := builder.Expr("ALTER TABLE ")
		e.WriteExpr(key.Table)
		e.WriteString(" ADD PRIMARY KEY ")
		e.WriteGroup(func(e *builder.Ex) {
			e.WriteExpr(key.Columns)
		})
		e.WriteEnd()
		return e
	}

	e := builder.Expr("CREATE ")
	if key.IsUnique {
		e.WriteString("UNIQUE ")
	}
	e.WriteString("INDEX ")

	e.WriteString(key.Table.Name)
	e.WriteString("_")
	e.WriteString(key.Name)

	e.WriteString(" ON ")
	e.WriteExpr(key.Table)

	if map[string]bool{
		"BTREE":   true,
		"HASH":    true,
		"SPATIAL": true,
		"GIST":    true,
	}[strings.ToUpper(key.Method)] {
		e.WriteString(" USING ")

		if key.Method == "SPATIAL" {
			e.WriteString("GIST")
		} else {
			e.WriteString(key.Method)
		}
	}

	e.WriteByte(' ')
	e.WriteGroup(func(e *builder.Ex) {
		e.WriteExpr(key.Columns)
	})

	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) DropIndex(key *builder.Key) builder.SqlExpr {
	if key.IsPrimary() {
		e := builder.Expr("ALTER TABLE ")
		e.WriteExpr(key.Table)
		e.WriteString(" DROP CONSTRAINT ")
		e.WriteString(key.Table.Name)
		e.WriteString("_pkey")
		e.WriteEnd()
		return e
	}
	e := builder.Expr("DROP ")

	e.WriteString("INDEX ")
	e.WriteString(key.Table.Name)
	e.WriteByte('_')
	e.WriteString(key.Name)

	return e
}

func (c *PostgreSQLConnector) CreateTableIsNotExists(t *builder.Table) (exprs []builder.SqlExpr) {
	expr := builder.Expr("CREATE TABLE IF NOT EXISTS ")
	expr.WriteString(t.Name)
	expr.WriteByte(' ')
	expr.WriteGroup(func(e *builder.Ex) {
		if t.Columns.IsNil() {
			return
		}

		t.Columns.Range(func(col *builder.Column, idx int) {
			if idx > 0 {
				e.WriteByte(',')
			}
			e.WriteByte('\n')
			e.WriteByte('\t')

			e.WriteExpr(col)
			e.WriteByte(' ')
			e.WriteExpr(c.DataType(col.ColumnType))
		})

		t.Keys.Range(func(key *builder.Key, idx int) {
			if key.IsPrimary() {
				e.WriteByte(',')
				e.WriteByte('\n')
				e.WriteByte('\t')
				e.WriteString("PRIMARY KEY ")
				e.WriteGroup(func(e *builder.Ex) {
					e.WriteExpr(key.Columns)
				})
			}
		})

		expr.WriteByte('\n')
	})

	expr.WriteEnd()
	exprs = append(exprs, expr)

	t.Keys.Range(func(key *builder.Key, idx int) {
		if !key.IsPrimary() {
			exprs = append(exprs, c.AddIndex(key))
		}
	})

	return
}

func (c *PostgreSQLConnector) DropTable(t *builder.Table) builder.SqlExpr {
	e := builder.Expr("DROP TABLE ")
	e.WriteString(t.Name)
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) TruncateTable(t *builder.Table) builder.SqlExpr {
	e := builder.Expr("TRUNCATE TABLE ")
	e.WriteString(t.Name)
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) AddColumn(col *builder.Column) builder.SqlExpr {
	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)
	e.WriteString(" ADD COLUMN ")
	e.WriteExpr(col)
	e.WriteByte(' ')
	e.WriteExpr(c.DataType(col.ColumnType))
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) ModifyColumn(col *builder.Column) builder.SqlExpr {
	if col.AutoIncrement {
		return nil
	}

	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)

	e.WriteString(" ALTER COLUMN ")
	e.WriteExpr(col)
	e.WriteString(" TYPE ")
	e.WriteString(c.dataType(col.ColumnType.Type, col.ColumnType))

	{
		e.WriteString(", ALTER COLUMN ")
		e.WriteExpr(col)
		if !col.Null {
			e.WriteString(" SET NOT NULL")
		} else {
			e.WriteString(" DROP NOT NULL")
		}
	}

	{
		e.WriteString(", ALTER COLUMN ")
		e.WriteExpr(col)
		if col.Default != nil {
			e.WriteString(" SET DEFAULT ")
			e.WriteString(*col.Default)
		} else {
			e.WriteString(" DROP DEFAULT")
		}
	}

	e.WriteEnd()

	return e
}

func (c *PostgreSQLConnector) DropColumn(col *builder.Column) builder.SqlExpr {
	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)
	e.WriteString(" DROP COLUMN ")
	e.WriteExpr(col)
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) DataType(columnType *builder.ColumnType) builder.SqlExpr {
	return builder.Expr(c.dataType(columnType.Type, columnType) + c.dataTypeModify(columnType))
}

func (c *PostgreSQLConnector) dataType(typ reflect.Type, columnType *builder.ColumnType) string {
	if columnType.GetDataType != nil {
		return columnType.GetDataType(c.DriverName())
	}

	switch typ.Kind() {
	case reflect.Ptr:
		return c.dataType(typ.Elem(), columnType)
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		if columnType.AutoIncrement {
			return "serial"
		}
		return "integer"
	case reflect.Int64, reflect.Uint64:
		if columnType.AutoIncrement {
			return "bigserial"
		}
		return "bigint"
	case reflect.Float64:
		return "double precision"
	case reflect.Float32:
		return "real"
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return "bytea"
		}
	case reflect.String:
		size := columnType.Length
		if size == 0 {
			size = 255
		}
		if size < 65535/3 {
			return "varchar" + sizeModifier(size, 0)
		}
		return "text"
	}

	switch typ.Name() {
	case "NullInt64":
		return "bigint"
	case "NullFloat64":
		return "double precision"
	case "NullBool":
		return "boolean"
	case "Time", "NullTime":
		return "timestamp with time zone"
	}

	panic(fmt.Errorf("unsupport type %s", typ))
}

func (c *PostgreSQLConnector) dataTypeModify(columnType *builder.ColumnType) string {
	buf := bytes.NewBuffer(nil)

	if !columnType.Null {
		buf.WriteString(" NOT NULL")
	}

	if columnType.Default != nil {
		buf.WriteString(" DEFAULT ")
		buf.WriteString(*columnType.Default)
	}

	return buf.String()
}

func sizeModifier(length uint64, decimal uint64) string {
	if length > 0 {
		size := strconv.FormatUint(length, 10)
		if decimal > 0 {
			return "(" + size + "," + strconv.FormatUint(decimal, 10) + ")"
		}
		return "(" + size + ")"
	}
	return ""
}
