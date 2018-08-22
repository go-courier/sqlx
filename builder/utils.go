package builder

import (
	"database/sql/driver"
	"reflect"
	"regexp"

	"github.com/go-courier/reflectx"
)

var queryRegexp = regexp.MustCompile(`(\$\d+)|\?`)

func FlattenArgs(query string, args ...interface{}) (finalQuery string, finalArgs []interface{}) {
	index := 0
	finalQuery = queryRegexp.ReplaceAllStringFunc(query, func(i string) string {
		arg := args[index]
		index++

		if canExpr, ok := arg.(SqlExpr); ok {
			e := canExpr.Expr()
			resolvedQuery, resolvedArgs := FlattenArgs(e.Query, e.Args...)
			finalArgs = append(finalArgs, resolvedArgs...)
			return resolvedQuery
		}

		if _, isValuer := arg.(driver.Valuer); !isValuer {
			typ := reflect.TypeOf(arg)
			if !reflectx.IsBytes(typ) && typ.Kind() == reflect.Slice {
				sliceRv := reflect.ValueOf(arg)
				length := sliceRv.Len()
				for i := 0; i < length; i++ {
					finalArgs = append(finalArgs, sliceRv.Index(i).Interface())
				}
				return HolderRepeat(length)
			}
		}
		finalArgs = append(finalArgs, arg)
		return i
	})
	return
}
