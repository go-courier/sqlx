package sqlx

import (
	"fmt"
)

func NewSqlError(tpe sqlErrType, msg string) *SqlError {
	return &SqlError{
		Type: tpe,
		Msg:  msg,
	}
}

type SqlError struct {
	Type sqlErrType
	Msg  string
}

func (e *SqlError) Error() string {
	return fmt.Sprintf("Sqlx [%s] %s", e.Type, e.Msg)
}

type sqlErrType string

var (
	sqlErrTypeInvalidScanTarget sqlErrType = "InvalidScanTarget"
	sqlErrTypeNotFound          sqlErrType = "NotFound"
	sqlErrTypeSelectShouldOne   sqlErrType = "SelectShouldOne"
	sqlErrTypeConflict          sqlErrType = "Conflict"
)

var DuplicateEntryErrNumber uint16 = 1062

func DBErr(err error) *dbErr {
	return &dbErr{
		err: err,
	}
}

type dbErr struct {
	err error

	errDefault  error
	errNotFound error
	errConflict error
}

func (r dbErr) WithNotFound(err error) *dbErr {
	r.errNotFound = err
	return &r
}

func (r dbErr) WithDefault(err error) *dbErr {
	r.errDefault = err
	return &r
}

func (r dbErr) WithConflict(err error) *dbErr {
	r.errConflict = err
	return &r
}

func (r *dbErr) IsNotFound() bool {
	if sqlErr, ok := r.err.(*SqlError); ok {
		return sqlErr.Type == sqlErrTypeNotFound
	}
	return false
}

func (r *dbErr) IsConflict() bool {
	if sqlErr, ok := r.err.(*SqlError); ok {
		return sqlErr.Type == sqlErrTypeConflict
	}
	return false
}

func (r *dbErr) Err() error {
	if r.err == nil {
		return nil
	}
	if sqlErr, ok := r.err.(*SqlError); ok {
		switch sqlErr.Type {
		case sqlErrTypeNotFound:
			if r.errNotFound != nil {
				return r.errNotFound
			}
		case sqlErrTypeConflict:
			if r.errConflict != nil {
				return r.errConflict
			}
		}
		if r.errDefault != nil {
			return r.errDefault
		}
	}
	return r.err
}
