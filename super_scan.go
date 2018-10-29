package sqlx

import (
	"database/sql"
	"reflect"

	"github.com/go-courier/sqlx/builder"
	"github.com/go-courier/sqlx/nullable"
)

func Scan(rows *sql.Rows, v interface{}) error {
	if rows == nil {
		return nil
	}

	defer rows.Close()

	if scanner, ok := v.(sql.Scanner); ok {
		for rows.Next() {
			if scanErr := rows.Scan(scanner); scanErr != nil {
				return scanErr
			}
		}
	} else {
		modelType := reflect.TypeOf(v)
		if modelType.Kind() != reflect.Ptr {
			return NewSqlError(sqlErrTypeInvalidScanTarget, "can not scan to a none pointer variable")
		}

		modelType = modelType.Elem()

		isSlice := false
		if modelType.Kind() == reflect.Slice {
			modelType = modelType.Elem()
			isSlice = true
		}

		if modelType.Kind() == reflect.Struct || isSlice {
			columns, getErr := rows.Columns()
			if getErr != nil {
				return getErr
			}

			rv := reflect.Indirect(reflect.ValueOf(v))

			rowLength := 0

			for rows.Next() {
				if !isSlice && rowLength > 1 {
					return NewSqlError(sqlErrTypeSelectShouldOne, "more than one records found, but only one")
				}

				rowLength++
				length := len(columns)
				dest := make([]interface{}, length)
				itemRv := rv

				if isSlice {
					itemRv = reflect.New(modelType).Elem()
				}

				destIndexes := make(map[int]bool, length)

				builder.ForEachStructFieldValue(itemRv, func(structFieldValue reflect.Value, structField reflect.StructField, columnName string, tagValue string) {
					idx := stringIndexOf(columns, columnName)
					if idx >= 0 {
						dest[idx] = structFieldValue.Addr().Interface()
						destIndexes[idx] = true
					}
				})

				for index := range dest {
					if !destIndexes[index] {
						placeholder := emptyScanner(0)
						dest[index] = &placeholder
					} else {
						dest[index] = nullable.NewNullIgnoreScanner(dest[index])
					}
				}

				if scanErr := rows.Scan(dest...); scanErr != nil {
					return scanErr
				}

				if isSlice {
					rv.Set(reflect.Append(rv, itemRv))
				}
			}

			if !isSlice && rowLength == 0 {
				return NewSqlError(sqlErrTypeNotFound, "record is not found")
			}
		} else {
			for rows.Next() {
				if scanErr := rows.Scan(v); scanErr != nil {
					return scanErr
				}
			}
		}
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

type emptyScanner int

var _ interface {
	sql.Scanner
} = (*emptyScanner)(nil)

func (e *emptyScanner) Scan(value interface{}) error {
	return nil
}

func stringIndexOf(slice []string, target string) int {
	for idx, item := range slice {
		if item == target {
			return idx
		}
	}
	return -1
}
