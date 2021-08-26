package scanner

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/sqlx/v2/scanner/nullable"
	reflectx "github.com/go-courier/x/reflect"
)

type RowScanner interface {
	Scan(values ...interface{}) error
}

type WithColumnReceivers interface {
	ColumnReceivers() map[string]interface{}
}

func scanTo(ctx context.Context, rows *sql.Rows, v interface{}) error {
	tpe := reflect.TypeOf(v)

	if tpe.Kind() != reflect.Ptr {
		return fmt.Errorf("scanTo target must be a ptr value, but got %T", v)
	}

	if s, ok := v.(sql.Scanner); ok {
		return rows.Scan(s)
	}

	tpe = reflectx.Deref(tpe)

	switch tpe.Kind() {
	case reflect.Struct:
		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		n := len(columns)
		if n < 1 {
			return nil
		}

		dest := make([]interface{}, n)
		holder := placeholder()

		if withColumnReceivers, ok := v.(WithColumnReceivers); ok {
			columnReceivers := withColumnReceivers.ColumnReceivers()

			for i, columnName := range columns {
				if cr, ok := columnReceivers[strings.ToLower(columnName)]; ok {
					dest[i] = nullable.NewNullIgnoreScanner(cr)
				} else {
					dest[i] = holder
				}
			}

			return rows.Scan(dest...)
		}

		columnIndexes := map[string]int{}

		for i, columnName := range columns {
			columnIndexes[strings.ToLower(columnName)] = i
			dest[i] = holder
		}

		builder.ForEachStructFieldValue(ctx, v, func(sf *builder.StructFieldValue) {
			if sf.TableName != "" {
				if i, ok := columnIndexes[sf.TableName+"__"+sf.Field.Name]; ok && i > -1 {
					dest[i] = nullable.NewNullIgnoreScanner(sf.Value.Addr().Interface())
				}
			}

			if i, ok := columnIndexes[sf.Field.Name]; ok && i > -1 {
				dest[i] = nullable.NewNullIgnoreScanner(sf.Value.Addr().Interface())
			}
		})

		return rows.Scan(dest...)
	default:
		return rows.Scan(nullable.NewNullIgnoreScanner(v))
	}
}

func placeholder() sql.Scanner {
	p := emptyScanner(0)
	return &p
}

type emptyScanner int

func (e *emptyScanner) Scan(value interface{}) error {
	return nil
}
