package sqlx

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/go-courier/sqlx/builder"
	_ "github.com/go-courier/sqlx/loggerdriver"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

var ErrNotTx = errors.New("db is not *sql.Tx")
var ErrNotDB = errors.New("db is not *sql.DB")

func MustOpen(driverName string, dataSourceName string) *DB {
	db, err := Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}
	return db
}

func Open(driverName string, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{SqlExecutor: db}, nil
}

type SqlExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type DB struct {
	SqlExecutor
}

func (d *DB) ExecExpr(expr builder.SqlExpr) (sql.Result, error) {
	e := expr.Expr()
	if e == nil {
		return nil, nil
	}
	if e.Err != nil {
		return nil, e.Err
	}
	result, err := d.Exec(e.Query, e.Args...)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == DuplicateEntryErrNumber {
			return nil, NewSqlError(sqlErrTypeConflict, mysqlErr.Error())
		}
		return nil, err
	}
	return result, nil
}

func (d *DB) QueryExprAndScan(expr builder.SqlExpr, v interface{}) error {
	rows, err := d.QueryExpr(expr)
	if err != nil {
		return err
	}
	return Scan(rows, v)
}

func (d *DB) QueryExpr(expr builder.SqlExpr) (*sql.Rows, error) {
	e := expr.Expr()
	if e == nil {
		return nil, nil
	}
	if e.Err != nil {
		return nil, e.Err
	}
	return d.Query(e.Query, e.Args...)
}

func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	query, args = builder.FlattenArgs(query, args...)
	return d.SqlExecutor.Query(query, args...)
}

func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	query, args = builder.FlattenArgs(query, args...)
	return d.SqlExecutor.Exec(query, args...)
}

func (d *DB) IsTx() bool {
	_, ok := d.SqlExecutor.(*sql.Tx)
	return ok
}

func (d *DB) Begin() (*DB, error) {
	if d.IsTx() {
		return nil, ErrNotDB
	}
	db, err := d.SqlExecutor.(*sql.DB).Begin()
	if err != nil {
		return nil, err
	}
	return &DB{
		SqlExecutor: db,
	}, nil
}

func (d *DB) Commit() error {
	if !d.IsTx() {
		return ErrNotTx
	}
	return d.SqlExecutor.(*sql.Tx).Commit()
}

func (d *DB) Rollback() error {
	if !d.IsTx() {
		return ErrNotTx
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

type Task func(db *DB) error

func (task Task) Run(db *DB) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %s; calltrace:%s", fmt.Sprint(e), string(debug.Stack()))
		}
	}()
	return task(db)
}

func NewTasks(db *DB) *Tasks {
	return &Tasks{
		db: db,
	}
}

type Tasks struct {
	db    *DB
	tasks []Task
}

func (tasks Tasks) With(task ...Task) *Tasks {
	tasks.tasks = append(tasks.tasks, task...)
	return &tasks
}

func (tasks *Tasks) Do() (err error) {
	if len(tasks.tasks) == 0 {
		return nil
	}

	db := tasks.db
	inTxScope := false

	if !db.IsTx() {
		db, err = db.Begin()
		if err != nil {
			return err
		}
		inTxScope = true
	}

	for _, task := range tasks.tasks {
		if runErr := task.Run(db); runErr != nil {
			if inTxScope {
				// err will bubble up，just handle and rollback in outermost layer
				logrus.Errorf("SQL FAILED: %s", runErr.Error())
				if rollBackErr := db.Rollback(); rollBackErr != nil {
					logrus.Errorf("ROLLBACK FAILED: %s", rollBackErr.Error())
					err = rollBackErr
					return
				}
			}
			return runErr
		}
	}

	if inTxScope {
		if commitErr := db.Commit(); commitErr != nil {
			logrus.Errorf("TRANSACTION COMMIT FAILED: %s", commitErr.Error())
			return commitErr
		}
	}

	return nil
}
