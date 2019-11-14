package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-courier/sqlx/v2/builder"
)

var ErrNotTx = errors.New("db is not *sql.Tx")
var ErrNotDB = errors.New("db is not *sql.DB")

type SqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type SqlxExecutor interface {
	SqlExecutor
	ExecExpr(expr builder.SqlExpr) (sql.Result, error)
	QueryExpr(expr builder.SqlExpr) (*sql.Rows, error)

	QueryExprAndScan(expr builder.SqlExpr, v interface{}) error
}

type Migrator interface {
	Migrate(ctx context.Context, db DBExecutor) error
}

type DBExecutor interface {
	SqlxExecutor

	// dialect of databases
	Dialect() builder.Dialect
	// return database which is connecting
	D() *Database
	// switch database schema
	WithSchema(schema string) DBExecutor
	// return table of the connecting database
	T(model builder.Model) *builder.Table

	Context() context.Context
	WithContext(ctx context.Context) DBExecutor
}

type MaybeTxExecutor interface {
	IsTx() bool
	BeginTx(*sql.TxOptions) (DBExecutor, error)
	Begin() (DBExecutor, error)
	Commit() error
	Rollback() error
}

type DB struct {
	dialect builder.Dialect
	*Database
	SqlExecutor
	ctx context.Context
}

func (d *DB) WithContext(ctx context.Context) DBExecutor {
	dd := new(DB)
	*dd = *d
	dd.ctx = ctx
	return dd
}

func (d *DB) Context() context.Context {
	if d.ctx != nil {
		return d.ctx
	}
	return context.Background()
}

func (d DB) WithSchema(schema string) DBExecutor {
	d.Database = d.Database.WithSchema(schema)
	return &d
}

func (d *DB) Dialect() builder.Dialect {
	return d.dialect
}

func (d *DB) Migrate(ctx context.Context, db DBExecutor) error {
	if migrator, ok := d.dialect.(Migrator); ok {
		return migrator.Migrate(ctx, db)
	}
	return nil
}

func (d *DB) D() *Database {
	return d.Database
}

func (d *DB) ExecExpr(expr builder.SqlExpr) (sql.Result, error) {
	e := builder.ExprFrom(expr)
	if e.IsNil() {
		return nil, nil
	}
	if err := e.Err(); err != nil {
		return nil, err
	}
	e = e.Flatten().ReplaceValueHolder(d.dialect.BindVar)
	result, err := d.ExecContext(d.Context(), e.Query(), e.Args()...)
	if err != nil {
		if d.dialect.IsErrorConflict(err) {
			return nil, NewSqlError(sqlErrTypeConflict, err.Error())
		}
		return nil, err
	}
	return result, nil
}

func (d *DB) QueryExpr(expr builder.SqlExpr) (*sql.Rows, error) {
	e := builder.ExprFrom(expr)
	if e.IsNil() {
		return nil, nil
	}
	if err := e.Err(); err != nil {
		return nil, err
	}
	e = e.Flatten().ReplaceValueHolder(d.dialect.BindVar)
	return d.QueryContext(d.Context(), e.Query(), e.Args()...)
}

func (d *DB) QueryExprAndScan(expr builder.SqlExpr, v interface{}) error {
	rows, err := d.QueryExpr(expr)
	if err != nil {
		return err
	}
	return Scan(rows, v)
}

func (d *DB) IsTx() bool {
	_, ok := d.SqlExecutor.(*sql.Tx)
	return ok
}

func (d *DB) Begin() (DBExecutor, error) {
	return d.BeginTx(nil)
}

func (d *DB) BeginTx(opt *sql.TxOptions) (DBExecutor, error) {
	if d.IsTx() {
		return nil, ErrNotDB
	}
	db, err := d.SqlExecutor.(*sql.DB).BeginTx(d.Context(), opt)
	if err != nil {
		return nil, err
	}
	return &DB{
		Database:    d.Database,
		dialect:     d.dialect,
		SqlExecutor: db,
		ctx:         d.Context(),
	}, nil
}

func (d *DB) Commit() error {
	if !d.IsTx() {
		return ErrNotTx
	}
	if d.Context().Err() == context.Canceled {
		return context.Canceled
	}
	return d.SqlExecutor.(*sql.Tx).Commit()
}

func (d *DB) Rollback() error {
	if !d.IsTx() {
		return ErrNotTx
	}
	if d.Context().Err() == context.Canceled {
		return context.Canceled
	}
	return d.SqlExecutor.(*sql.Tx).Rollback()
}

func (d *DB) SetMaxOpenConns(n int) {
	d.SqlExecutor.(*sql.DB).SetMaxOpenConns(n)
}

func (d *DB) SetMaxIdleConns(n int) {
	d.SqlExecutor.(*sql.DB).SetMaxIdleConns(n)
}

func (d *DB) SetConnMaxLifetime(t time.Duration) {
	d.SqlExecutor.(*sql.DB).SetConnMaxLifetime(t)
}
