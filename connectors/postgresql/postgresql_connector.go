package postgresql

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	typex "github.com/go-courier/x/types"

	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/sqlx/v2/migration"
	"github.com/lib/pq"
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

func (c *PostgreSQLConnector) Connect(ctx context.Context) (driver.Conn, error) {
	d := c.Driver()

	conn, err := d.Open(dsn(c.Host, c.DBName, c.Extra))
	if err != nil {
		if c.IsErrorUnknownDatabase(err) {
			connectForCreateDB, err := d.Open(dsn(c.Host, "", c.Extra))
			if err != nil {
				return nil, err
			}
			if _, err := connectForCreateDB.(driver.ExecerContext).ExecContext(context.Background(), builder.ResolveExpr(c.CreateDatabase(c.DBName)).Query(), nil); err != nil {
				return nil, err
			}
			if err := connectForCreateDB.Close(); err != nil {
				return nil, err
			}
			return c.Connect(ctx)
		}
		return nil, err
	}
	for _, ex := range c.Extensions {
		if _, err := conn.(driver.ExecerContext).ExecContext(context.Background(), "CREATE EXTENSION IF NOT EXISTS "+ex+";", nil); err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func (PostgreSQLConnector) Driver() driver.Driver {
	return &PostgreSQLLoggingDriver{}
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

func (c *PostgreSQLConnector) Migrate(ctx context.Context, db sqlx.DBExecutor) error {
	output := migration.MigrationOutputFromContext(ctx)

	prevDB, err := dbFromInformationSchema(db)
	if err != nil {
		return err
	}

	d := db.D()
	dialect := db.Dialect()

	exec := func(expr builder.SqlExpr) error {
		if expr == nil || expr.IsNil() {
			return nil
		}

		if output != nil {
			_, _ = io.WriteString(output, builder.ResolveExpr(expr).Query())
			_, _ = io.WriteString(output, "\n")
			return nil
		}

		_, err := db.ExecExpr(expr)
		return err
	}

	if prevDB == nil {
		prevDB = &sqlx.Database{
			Name: d.Name,
		}
		if err := exec(dialect.CreateDatabase(d.Name)); err != nil {
			return err
		}
	}

	if d.Schema != "" {
		if err := exec(dialect.CreateSchema(d.Schema)); err != nil {
			return err
		}
		prevDB = prevDB.WithSchema(d.Schema)
	}

	for _, name := range d.Tables.TableNames() {
		table := d.Table(name)

		prevTable := prevDB.Table(name)

		if prevTable == nil {
			for _, expr := range dialect.CreateTableIsNotExists(table) {
				if err := exec(expr); err != nil {
					return err
				}
			}
			continue
		}

		exprList := table.Diff(prevTable, dialect)

		for _, expr := range exprList {
			if err := exec(expr); err != nil {
				return err
			}
		}
	}

	return nil
}

func (PostgreSQLConnector) DriverName() string {
	return "postgres"
}

func (PostgreSQLConnector) PrimaryKeyName() string {
	return "pkey"
}

func (PostgreSQLConnector) IsErrorUnknownDatabase(err error) bool {
	if e, ok := sqlx.UnwrapAll(err).(*pq.Error); ok && e.Code == "3D000" {
		return true
	}
	return false
}

func (PostgreSQLConnector) IsErrorConflict(err error) bool {
	if e, ok := sqlx.UnwrapAll(err).(*pq.Error); ok && e.Code == "23505" {
		return true
	}
	return false
}

func (c *PostgreSQLConnector) CreateDatabase(dbName string) builder.SqlExpr {
	e := builder.Expr("CREATE DATABASE ")
	e.WriteQuery(dbName)
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) CreateSchema(schema string) builder.SqlExpr {
	e := builder.Expr("CREATE SCHEMA IF NOT EXISTS ")
	e.WriteQuery(schema)
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) DropDatabase(dbName string) builder.SqlExpr {
	e := builder.Expr("DROP DATABASE IF EXISTS ")
	e.WriteQuery(dbName)
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) AddIndex(key *builder.Key) builder.SqlExpr {
	if key.IsPrimary() {
		e := builder.Expr("ALTER TABLE ")
		e.WriteExpr(key.Table)
		e.WriteQuery(" ADD PRIMARY KEY ")
		e.WriteExpr(key.Def.TableExpr(key.Table))
		e.WriteEnd()
		return e
	}

	e := builder.Expr("CREATE ")
	if key.IsUnique {
		e.WriteQuery("UNIQUE ")
	}
	e.WriteQuery("INDEX ")

	e.WriteQuery(key.Table.Name)
	e.WriteQuery("_")
	e.WriteQuery(key.Name)

	e.WriteQuery(" ON ")
	e.WriteExpr(key.Table)

	if m := strings.ToUpper(key.Method); m != "" {
		if m == "SPATIAL" {
			m = "GIST"
		}
		e.WriteQuery(" USING ")
		e.WriteQuery(m)
	}

	e.WriteQueryByte(' ')
	e.WriteExpr(key.Def.TableExpr(key.Table))

	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) DropIndex(key *builder.Key) builder.SqlExpr {
	if key.IsPrimary() {
		e := builder.Expr("ALTER TABLE ")
		e.WriteExpr(key.Table)
		e.WriteQuery(" DROP CONSTRAINT ")
		e.WriteExpr(key.Table)
		e.WriteQuery("_pkey")
		e.WriteEnd()
		return e
	}
	e := builder.Expr("DROP ")

	e.WriteQuery("INDEX IF EXISTS ")
	e.WriteExpr(key.Table)
	e.WriteQueryByte('_')
	e.WriteQuery(key.Name)
	e.WriteEnd()

	return e
}

func (c *PostgreSQLConnector) CreateTableIsNotExists(t *builder.Table) (exprs []builder.SqlExpr) {
	expr := builder.Expr("CREATE TABLE IF NOT EXISTS ")
	expr.WriteExpr(t)
	expr.WriteQueryByte(' ')
	expr.WriteGroup(func(e *builder.Ex) {
		if t.Columns.IsNil() {
			return
		}

		t.Columns.Range(func(col *builder.Column, idx int) {
			if col.DeprecatedActions != nil {
				return
			}

			if idx > 0 {
				e.WriteQueryByte(',')
			}
			e.WriteQueryByte('\n')
			e.WriteQueryByte('\t')

			e.WriteExpr(col)
			e.WriteQueryByte(' ')
			e.WriteExpr(c.DataType(col.ColumnType))
		})

		t.Keys.Range(func(key *builder.Key, idx int) {
			if key.IsPrimary() {
				e.WriteQueryByte(',')
				e.WriteQueryByte('\n')
				e.WriteQueryByte('\t')
				e.WriteQuery("PRIMARY KEY ")
				e.WriteExpr(key.Def.TableExpr(key.Table))
			}
		})

		expr.WriteQueryByte('\n')
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
	e := builder.Expr("DROP TABLE IF EXISTS ")
	e.WriteExpr(t)
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) TruncateTable(t *builder.Table) builder.SqlExpr {
	e := builder.Expr("TRUNCATE TABLE ")
	e.WriteExpr(t)
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) AddColumn(col *builder.Column) builder.SqlExpr {
	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)
	e.WriteQuery(" ADD COLUMN ")
	e.WriteExpr(col)
	e.WriteQueryByte(' ')
	e.WriteExpr(c.DataType(col.ColumnType))
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) RenameColumn(col *builder.Column, target *builder.Column) builder.SqlExpr {
	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)
	e.WriteQuery(" RENAME COLUMN ")
	e.WriteExpr(col)
	e.WriteQuery(" TO ")
	e.WriteExpr(target)
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) ModifyColumn(col *builder.Column, prev *builder.Column) builder.SqlExpr {
	if col.AutoIncrement {
		return nil
	}

	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)

	dbDataType := c.dataType(col.ColumnType.Type, col.ColumnType)
	prevDbDataType := c.dataType(prev.ColumnType.Type, prev.ColumnType)

	isFirstSub := true
	isEmpty := true

	prepareAppendSubCmd := func() {
		if !isFirstSub {
			e.WriteQueryByte(',')
		}
		isFirstSub = false
		isEmpty = false
	}

	if dbDataType != prevDbDataType {
		prepareAppendSubCmd()

		e.WriteQuery(" ALTER COLUMN ")
		e.WriteExpr(col)
		e.WriteQuery(" TYPE ")
		e.WriteQuery(dbDataType)

		e.WriteQuery(" /* FROM ")
		e.WriteQuery(prevDbDataType)
		e.WriteQuery(" */")
	}

	if col.Null != prev.Null {
		prepareAppendSubCmd()

		e.WriteQuery(" ALTER COLUMN ")
		e.WriteExpr(col)
		if !col.Null {
			e.WriteQuery(" SET NOT NULL")
		} else {
			e.WriteQuery(" DROP NOT NULL")
		}
	}

	defaultValue := normalizeDefaultValue(col.Default, dbDataType)
	prevDefaultValue := normalizeDefaultValue(prev.Default, prevDbDataType)

	if defaultValue != prevDefaultValue {
		prepareAppendSubCmd()

		e.WriteQuery(" ALTER COLUMN ")
		e.WriteExpr(col)
		if col.Default != nil {
			e.WriteQuery(" SET DEFAULT ")
			e.WriteQuery(defaultValue)

			e.WriteQuery(" /* FROM ")
			e.WriteQuery(prevDefaultValue)
			e.WriteQuery(" */")
		} else {
			e.WriteQuery(" DROP DEFAULT")
		}
	}

	if isEmpty {
		return nil
	}

	e.WriteEnd()

	return e
}

func (c *PostgreSQLConnector) DropColumn(col *builder.Column) builder.SqlExpr {
	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)
	e.WriteQuery(" DROP COLUMN ")
	e.WriteQuery(col.Name)
	e.WriteEnd()
	return e
}

func (c *PostgreSQLConnector) DataType(columnType *builder.ColumnType) builder.SqlExpr {
	dbDataType := dealias(c.dbDataType(columnType.Type, columnType))
	return builder.Expr(dbDataType + autocompleteSize(dbDataType, columnType) + c.dataTypeModify(columnType, dbDataType))
}

func (c *PostgreSQLConnector) dataType(typ typex.Type, columnType *builder.ColumnType) string {
	dbDataType := dealias(c.dbDataType(columnType.Type, columnType))
	return dbDataType + autocompleteSize(dbDataType, columnType)
}

func (c *PostgreSQLConnector) dbDataType(typ typex.Type, columnType *builder.ColumnType) string {
	if columnType.DataType != "" {
		return columnType.DataType
	}

	if rv, ok := typex.TryNew(typ); ok {
		if dtd, ok := rv.Interface().(builder.DataTypeDescriber); ok {
			return dtd.DataType(c.DriverName())
		}
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
		if size < 65535/3 {
			return "varchar"
		}
		return "text"
	}

	switch typ.Name() {
	case "Hstore":
		return "hstore"
	case "ByteaArray":
		return c.dataType(typex.FromRType(reflect.TypeOf(pq.ByteaArray{[]byte("")}[0])), columnType) + "[]"
	case "BoolArray":
		return c.dataType(typex.FromRType(reflect.TypeOf(pq.BoolArray{true}[0])), columnType) + "[]"
	case "Float64Array":
		return c.dataType(typex.FromRType(reflect.TypeOf(pq.Float64Array{0}[0])), columnType) + "[]"
	case "Int64Array":
		return c.dataType(typex.FromRType(reflect.TypeOf(pq.Int64Array{0}[0])), columnType) + "[]"
	case "StringArray":
		return c.dataType(typex.FromRType(reflect.TypeOf(pq.StringArray{""}[0])), columnType) + "[]"
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

func (c *PostgreSQLConnector) dataTypeModify(columnType *builder.ColumnType, dataType string) string {
	buf := bytes.NewBuffer(nil)

	if !columnType.Null {
		buf.WriteString(" NOT NULL")
	}

	if columnType.Default != nil {
		buf.WriteString(" DEFAULT ")
		buf.WriteString(normalizeDefaultValue(columnType.Default, dataType))
	}

	return buf.String()
}

func normalizeDefaultValue(defaultValue *string, dataType string) string {
	if defaultValue == nil {
		return ""
	}

	dv := *defaultValue

	if dv[0] == '\'' {
		if strings.Contains(dv, "'::") {
			return dv
		}
		return dv + "::" + dataType
	}

	_, err := strconv.ParseFloat(dv, 64)
	if err == nil {
		return "'" + dv + "'::" + dataType
	}

	return dv
}

func autocompleteSize(dataType string, columnType *builder.ColumnType) string {
	switch dataType {
	case "character varying", "character":
		size := columnType.Length
		if size == 0 {
			size = 255
		}
		return sizeModifier(size, columnType.Decimal)
	case "decimal", "numeric", "real", "double precision":
		if columnType.Length > 0 {
			return sizeModifier(columnType.Length, columnType.Decimal)
		}
	}
	return ""
}

func dealias(dataType string) string {
	switch dataType {
	case "varchar":
		return "character varying"
	case "timestamp":
		return "timestamp without time zone"
	}
	return dataType
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
