package mysqlconnector

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/go-courier/sqlx"
	"github.com/go-courier/sqlx/builder"
	"github.com/go-courier/sqlx/migration"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"log"
	"reflect"
	"strconv"
)

var (
	ErrNumberUnknownDatabase uint16 = 1049
	ErrNumberDuplicateEntry  uint16 = 1062
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
	return host + "/" + dbName + "?" + extra
}

func (c MysqlConnector) WithDBName(dbName string) driver.Connector {
	c.DBName = dbName
	return &c
}

func (c *MysqlConnector) Migrate(db *sqlx.DB,database *sqlx.Database, opts *migration.MigrationOpts) error {
	if opts == nil {
		opts = &migration.MigrationOpts{}
	}

	prevDB := DBFromInformationSchema(db, database.Name, database.Tables.TableNames()...)

	if prevDB == nil {
		prevDB = &sqlx.Database{
			Name: database.Name,
		}
		if _, err := db.ExecExpr(db.CreateDatabaseIfNotExists(database.Name)); err != nil {
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
			if !expr.IsNil() {
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

func (c *MysqlConnector) Connect(ctx context.Context) (driver.Conn, error) {
	d := c.Driver()
	conn, err := d.Open(dsn(c.Host, c.DBName, c.Extra))
	if err != nil {
		if c.IsErrorUnknownDatabase(err) {
			conn, err := d.Open(dsn(c.Host, "", c.Extra))
			if err != nil {
				return nil, err
			}
			stmt, _ := conn.Prepare(c.CreateDatabaseIfNotExists(c.DBName).Expr().Query())
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
	return conn, nil
}

func (MysqlConnector) Driver() driver.Driver {
	return &MySqlLoggingDriver{Driver: &mysql.MySQLDriver{}, Logger: logrus.StandardLogger()}
}

func (MysqlConnector) BindVar(i int) string {
	return "?"
}

func (MysqlConnector) DriverName() string {
	return "mysql"
}

func (c MysqlConnector) IsErrorUnknownDatabase(err error) bool {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == ErrNumberUnknownDatabase {
		return true
	}
	return false
}

func (c MysqlConnector) IsErrorConflict(err error) bool {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == ErrNumberDuplicateEntry {
		return true
	}
	return false
}

func (c *MysqlConnector) CreateDatabaseIfNotExists(dbName string) builder.SqlExpr {
	e := builder.Expr("CREATE DATABASE IF NOT EXISTS ")
	e.WriteString(dbName)
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

func (c *MysqlConnector) CreateTableIsNotExists(t *builder.Table) (exprs []builder.SqlExpr) {
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

	t.Keys.Range(func(key *builder.Key, idx int) {
		if !key.IsPrimary() {
			exprs = append(exprs, c.AddIndex(key))
		}
	})

	return
}

func (c *MysqlConnector) DropTable(t *builder.Table) builder.SqlExpr {
	e := builder.Expr("DROP TABLE ")
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

func (c *MysqlConnector) ModifyColumn(col *builder.Column) builder.SqlExpr {
	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)
	e.WriteString(" MODIFY COLUMN ")
	e.WriteExpr(col)
	e.WriteByte(' ')
	e.WriteExpr(c.DataType(col.ColumnType))
	e.WriteEnd()
	return e
}

func (c *MysqlConnector) DropColumn(col *builder.Column) builder.SqlExpr {
	e := builder.Expr("ALTER TABLE ")
	e.WriteExpr(col.Table)
	e.WriteString(" DROP COLUMN ")
	e.WriteExpr(col)
	e.WriteEnd()
	return e
}

func (c *MysqlConnector) DataType(columnType *builder.ColumnType) builder.SqlExpr {
	return builder.Expr(c.dataType(columnType.Type, columnType) + c.dataTypeModify(columnType))
}

func (c *MysqlConnector) dataType(typ reflect.Type, columnType *builder.ColumnType) string {
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
		return "float" + sizeModifier(columnType.Length, columnType.Decimal)
	case reflect.Float64:
		return "double" + sizeModifier(columnType.Length, columnType.Decimal)
	case reflect.String:
		size := columnType.Length
		if size == 0 {
			size = 255
		}
		if size < 65535/3 {
			return "varchar" + sizeModifier(size, 0)
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
