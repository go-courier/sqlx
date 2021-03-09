package mysqlconnector

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/sqlx/v2/migration"
	"github.com/go-sql-driver/mysql"
)

var _ interface {
	driver.Connector
	builder.Dialect
} = (*MysqlConnector)(nil)

type MysqlConnector struct {
	Host    string
	DBName  string
	Extra   string
	Engine  string
	Charset string
}

func dsn(host string, dbName string, extra string) string {
	if extra != "" {
		extra = "?" + extra
	}
	return host + "/" + dbName + extra
}

func (c MysqlConnector) WithDBName(dbName string) driver.Connector {
	c.DBName = dbName
	return &c
}

func (c *MysqlConnector) Migrate(ctx context.Context, db sqlx.DBExecutor) error {
	output := migration.MigrationOutputFromContext(ctx)

	// mysql without schema
	d := db.D().WithSchema("")
	dialect := db.Dialect()

	prevDB := dbFromInformationSchema(db)

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

	for _, name := range d.Tables.TableNames() {
		table := d.Tables.Table(name)
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

func (c *MysqlConnector) Connect(ctx context.Context) (driver.Conn, error) {
	d := &MySqlLoggingDriver{}

	connector, err := d.OpenConnector(dsn(c.Host, c.DBName, c.Extra))
	if err != nil {
		return nil, err
	}

	conn, err := connector.Connect(ctx)
	if err != nil {
		if c.IsErrorUnknownDatabase(err) {
			conn, err := c.WithDBName("").Connect(ctx)
			if err != nil {
				return nil, err
			}
			if _, err := conn.(driver.ExecerContext).ExecContext(context.Background(), builder.ResolveExpr(c.CreateDatabase(c.DBName)).Query(), nil); err != nil {
				return nil, err
			}
			if err := conn.Close(); err != nil {
				return nil, err
			}
			return c.Connect(ctx)
		}
		return nil, err
	}
	return conn, nil
}

func (MysqlConnector) Driver() driver.Driver {
	return &MySqlLoggingDriver{}
}

func (MysqlConnector) DriverName() string {
	return "mysql"
}

func (MysqlConnector) PrimaryKeyName() string {
	return "primary"
}

func (c MysqlConnector) IsErrorUnknownDatabase(err error) bool {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1049 {
		return true
	}
	return false
}

func (c MysqlConnector) IsErrorConflict(err error) bool {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
		return true
	}
	return false
}

func (c *MysqlConnector) CreateDatabase(dbName string) builder.SqlExpr {
	e := builder.Expr("CREATE DATABASE ")
	e.WriteString(dbName)
	e.WriteEnd()
	return e
}

func (c *MysqlConnector) CreateSchema(schema string) builder.SqlExpr {
	e := builder.Expr("CREATE SCHEMA ")
	e.WriteString(schema)
	e.WriteEnd()
	return e
}

func (c *MysqlConnector) DropDatabase(dbName string) builder.SqlExpr {
	e := builder.Expr("DROP DATABASE ")
	e.WriteString(dbName)
	e.WriteEnd()
	return e
}

func (c *MysqlConnector) AddIndex(key *builder.Key) builder.SqlExpr {
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
	if key.Method == "SPATIAL" {
		e.WriteString("SPATIAL ")
	} else if key.IsUnique {
		e.WriteString("UNIQUE ")
	}
	e.WriteString("INDEX ")

	e.WriteString(key.Name)

	e.WriteString(" ON ")
	e.WriteExpr(key.Table)
	e.WriteByte(' ')
	e.WriteGroup(func(e *builder.Ex) {
		e.WriteExpr(key.Columns)
	})

	if key.Method == "BTREE" || key.Method == "HASH" {
		e.WriteString(" USING ")
		e.WriteString(key.Method)
	}

	e.WriteEnd()
	return e
}

func (c *MysqlConnector) DropIndex(key *builder.Key) builder.SqlExpr {
	if key.IsPrimary() {
		e := builder.Expr("ALTER TABLE ")
		e.WriteExpr(key.Table)
		e.WriteString(" DROP PRIMARY KEY")
		e.WriteEnd()
		return e
	}
	e := builder.Expr("DROP ")

	e.WriteString("INDEX ")
	e.WriteString(key.Name)

	e.WriteString(" ON ")
	e.WriteExpr(key.Table)
	e.WriteEnd()

	return e
}

func (c *MysqlConnector) CreateTableIsNotExists(table *builder.Table) (exprs []builder.SqlExpr) {
	expr := builder.Expr("CREATE TABLE IF NOT EXISTS ")
	expr.WriteExpr(table)
	expr.WriteByte(' ')
	expr.WriteGroup(func(e *builder.Ex) {
		if table.Columns.IsNil() {
			return
		}

		table.Columns.Range(func(col *builder.Column, idx int) {
			if col.DeprecatedActions != nil {
				return
			}

			if idx > 0 {
				e.WriteByte(',')
			}
			e.WriteByte('\n')
			e.WriteByte('\t')

			e.WriteExpr(col)
			e.WriteByte(' ')
			e.WriteExpr(c.DataType(col.ColumnType))
		})

		table.Keys.Range(func(key *builder.Key, idx int) {
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

	expr.WriteString(" ENGINE=")

	if c.Engine == "" {
		expr.WriteString("InnoDB")
	} else {
		expr.WriteString(c.Engine)
	}

	expr.WriteString(" CHARSET=")

	if c.Charset == "" {
		expr.WriteString("utf8mb4")
	} else {
		expr.WriteString(c.Charset)
	}

	expr.WriteEnd()
	exprs = append(exprs, expr)

	table.Keys.Range(func(key *builder.Key, idx int) {
		if !key.IsPrimary() {
			exprs = append(exprs, c.AddIndex(key))
		}
	})

	return
}

func (c *MysqlConnector) DropTable(t *builder.Table) builder.SqlExpr {
	e := builder.Expr("DROP TABLE IF EXISTS ")
	e.WriteString(t.Name)
	e.WriteEnd()
	return e
}

func (c *MysqlConnector) TruncateTable(t *builder.Table) builder.SqlExpr {
	e := builder.Expr("TRUNCATE TABLE ")
	e.WriteString(t.Name)
	e.WriteEnd()
	return e
}

func (c *MysqlConnector) AddColumn(col *builder.Column) builder.SqlExpr {
	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)
	e.WriteString(" ADD COLUMN ")
	e.WriteExpr(col)
	e.WriteByte(' ')
	e.WriteExpr(c.DataType(col.ColumnType))
	e.WriteEnd()
	return e
}

func (c *MysqlConnector) RenameColumn(col *builder.Column, target *builder.Column) builder.SqlExpr {
	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)
	e.WriteString(" CHANGE ")
	e.WriteExpr(col)
	e.WriteByte(' ')
	e.WriteExpr(target)
	e.WriteByte(' ')
	e.WriteExpr(c.DataType(target.ColumnType))
	e.WriteEnd()
	return e
}

func (c *MysqlConnector) ModifyColumn(col *builder.Column, prev *builder.Column) builder.SqlExpr {
	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)
	e.WriteString(" MODIFY COLUMN ")
	e.WriteExpr(col)
	e.WriteByte(' ')
	e.WriteExpr(c.DataType(col.ColumnType))

	e.WriteString(" /* FROM")
	e.WriteExpr(c.DataType(prev.ColumnType))
	e.WriteString(" */")

	e.WriteEnd()
	return e
}

func (c *MysqlConnector) DropColumn(col *builder.Column) builder.SqlExpr {
	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)
	e.WriteString(" DROP COLUMN ")
	e.WriteString(col.Name)
	e.WriteEnd()
	return e
}

func (c *MysqlConnector) DataType(columnType *builder.ColumnType) builder.SqlExpr {
	dbDataType := dealias(c.dbDataType(columnType.Type, columnType))
	return builder.Expr(dbDataType + autocompleteSize(dbDataType, columnType) + c.dataTypeModify(columnType))
}

func (c *MysqlConnector) dataType(typ reflect.Type, columnType *builder.ColumnType) string {
	dbDataType := dealias(c.dbDataType(typ, columnType))
	return dbDataType + autocompleteSize(dbDataType, columnType)
}

func (c *MysqlConnector) dbDataType(typ reflect.Type, columnType *builder.ColumnType) string {
	if columnType.GetDataType != nil {
		return columnType.GetDataType(c.DriverName())
	}

	switch typ.Kind() {
	case reflect.Ptr:
		return c.dataType(typ.Elem(), columnType)
	case reflect.Bool:
		return "boolean"
	case reflect.Int8:
		return "tinyint"
	case reflect.Uint8:
		return "tinyint unsigned"
	case reflect.Int16:
		return "smallint"
	case reflect.Uint16:
		return "smallint unsigned"
	case reflect.Int, reflect.Int32:
		return "int"
	case reflect.Uint, reflect.Uint32:
		return "int unsigned"
	case reflect.Int64:
		return "bigint"
	case reflect.Uint64:
		return "bigint unsigned"
	case reflect.Float32:
		return "float"
	case reflect.Float64:
		return "double"
	case reflect.String:
		size := columnType.Length
		if size < 65535/3 {
			return "varchar"
		}
		return "text"
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return "mediumblob"
		}
	}
	switch typ.Name() {
	case "NullInt64":
		return "bigint"
	case "NullFloat64":
		return "double"
	case "NullBool":
		return "tinyint"
	case "Time":
		return "datetime"
	}
	panic(fmt.Errorf("unsupport type %s", typ))
}

func (c *MysqlConnector) dataTypeModify(columnType *builder.ColumnType) string {
	buf := bytes.NewBuffer(nil)

	if !columnType.Null {
		buf.WriteString(" NOT NULL")
	}

	if columnType.AutoIncrement {
		buf.WriteString(" AUTO_INCREMENT")
	}

	if columnType.Default != nil {
		buf.WriteString(" DEFAULT ")
		buf.WriteString(*columnType.Default)
	}

	if columnType.OnUpdate != nil {
		buf.WriteString(" ON UPDATE ")
		buf.WriteString(*columnType.OnUpdate)
	}

	return buf.String()
}

func autocompleteSize(dataType string, columnType *builder.ColumnType) string {
	switch strings.ToLower(dataType) {
	case "varchar":
		size := columnType.Length
		if size == 0 {
			size = 255
		}
		return sizeModifier(size, columnType.Decimal)
	case "float", "double", "decimal":
		if columnType.Length > 0 {
			return sizeModifier(columnType.Length, columnType.Decimal)
		}
	}
	return ""
}

func dealias(dataType string) string {
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
