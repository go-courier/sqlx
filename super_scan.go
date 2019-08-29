package sqlx

import (
	"database/sql"
	"reflect"
	"strings"

	"github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/sqlx/v2/nullable"
)

type ScanIterator interface {
	// new a ptr value for scan
	New() interface{}
	// for receive scanned value
	Next(v interface{}) error
}

func Scan(rows *sql.Rows, v interface{}) error {
	if rows == nil {
		return nil
	}

	defer rows.Close()

	modelScanner, err := newModelScanner(v)
	if err != nil {
		return err
	}

	// simple value scan
	for rows.Next() {
		if modelScanner.direct {
			if scanErr := rows.Scan(modelScanner.v); scanErr != nil {
				return scanErr
			}
			continue
		}

		rv, err := modelScanner.New()
		if err != nil {
			return err
		}

		if scanErr := scanStruct(rows, rv); scanErr != nil {
			return scanErr
		}

		if err := modelScanner.Next(rv); err != nil {
			return err
		}
	}

	if !modelScanner.direct && !modelScanner.isSlice && modelScanner.count == 0 {
		return NewSqlError(sqlErrTypeNotFound, "record is not found")
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Make sure the query can be processed to completion with no errors.
	if err := rows.Close(); err != nil {
		return err
	}

	return nil
}

func newModelScanner(v interface{}) (*modelScanner, error) {
	si := &modelScanner{v: v}

	if _, ok := v.(sql.Scanner); ok {
		si.direct = true
	} else if scanIterator, ok := v.(ScanIterator); ok {
		si.scanIterator = scanIterator
	} else {
		modelType := reflect.TypeOf(v)

		if modelType.Kind() != reflect.Ptr {
			return nil, NewSqlError(sqlErrTypeInvalidScanTarget, "can not scan to a none pointer variable")
		}

		si.modelType = modelType.Elem()

		if si.modelType.Kind() == reflect.Slice {
			si.modelType = si.modelType.Elem()
			si.isSlice = true
		}

		si.rv = reflect.Indirect(reflect.ValueOf(v))
		si.direct = si.modelType.Kind() != reflect.Struct
	}

	return si, nil
}

type modelScanner struct {
	v            interface{}
	rv           reflect.Value
	direct       bool
	isSlice      bool
	count        int
	modelType    reflect.Type
	scanIterator ScanIterator
}

func (s *modelScanner) New() (reflect.Value, error) {
	if s.scanIterator != nil {
		rv := reflect.ValueOf(s.scanIterator.New())
		if rv.Kind() != reflect.Ptr {
			return reflect.Value{}, NewSqlError(sqlErrTypeInvalidScanTarget, "can not scan to a none pointer variable")
		}
		return rv.Elem(), nil
	}
	if s.isSlice {
		return reflect.New(s.modelType).Elem(), nil
	}
	return s.rv, nil
}

func (s *modelScanner) Next(rv reflect.Value) error {
	s.count++

	if s.scanIterator != nil {
		return s.scanIterator.Next(rv.Addr().Interface())
	}

	if s.isSlice {
		s.rv.Set(reflect.Append(s.rv, rv))
		return nil
	}

	s.rv.Set(rv)
	return nil
}

func scanStruct(rows *sql.Rows, rv reflect.Value) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	n := len(columns)
	dest := make([]interface{}, n)

	columnIndexes := map[string]int{}
	p := placeholder()

	for i, name := range columns {
		columnIndexes[strings.ToLower(name)] = i
		dest[i] = p
	}

	builder.ForEachStructFieldValue(rv, func(structFieldValue reflect.Value, structField reflect.StructField, columnName string, tagValue string) {
		if i, ok := columnIndexes[columnName]; ok && i > -1 {
			dest[i] = nullable.NewNullIgnoreScanner(structFieldValue.Addr().Interface())
		}
	})

	return rows.Scan(dest...)
}

func placeholder() *emptyScanner {
	p := emptyScanner(0)
	return &p
}

type emptyScanner int

func (e *emptyScanner) Scan(value interface{}) error {
	return nil
}
