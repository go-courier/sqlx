package builder

import (
	"fmt"
	"github.com/go-courier/reflectx"
	"go/ast"
	"reflect"
	"strings"
)

type FieldValues map[string]interface{}

func ForEachStructFieldValue(rv reflect.Value, fn func(structFieldValue reflect.Value, structField reflect.StructField, columnName string, tagValue string)) {
	rv = reflect.Indirect(rv)
	structType := rv.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if ast.IsExported(field.Name) {
			fieldValue := rv.Field(i)

			tagValue, exists := field.Tag.Lookup("db")
			if exists {
				if tagValue != "-" {
					fn(fieldValue, field, GetColumnName(field.Name, tagValue), tagValue)
				}
			} else if field.Anonymous {
				ForEachStructFieldValue(fieldValue, fn)
				continue
			}
		}
	}
}

func GetColumnName(fieldName, tagValue string) string {
	columnName := strings.Split(tagValue, ",")[0]
	if columnName == "" {
		return "f_" + strings.ToLower(fieldName)
	}
	return columnName
}

func ToMap(list []string) map[string]bool {
	m := make(map[string]bool, len(list))
	for _, fieldName := range list {
		m[fieldName] = true
	}
	return m
}

func FieldValuesFromStructBy(structValue interface{}, fieldNames []string) (fieldValues FieldValues) {
	fieldValues = FieldValues{}
	rv := reflect.Indirect(reflect.ValueOf(structValue))
	fieldMap := ToMap(fieldNames)
	ForEachStructFieldValue(rv, func(structFieldValue reflect.Value, structField reflect.StructField, columnName string, tagValue string) {
		if fieldMap != nil && fieldMap[structField.Name] {
			fieldValues[structField.Name] = structFieldValue.Interface()
		}
	})
	return fieldValues
}

func FieldValuesFromStructByNonZero(structValue interface{}, excludes ...string) (fieldValues FieldValues) {
	fieldValues = FieldValues{}
	rv := reflect.Indirect(reflect.ValueOf(structValue))
	fieldMap := ToMap(excludes)
	ForEachStructFieldValue(rv, func(structFieldValue reflect.Value, structField reflect.StructField, columnName string, tagValue string) {
		if !reflectx.IsEmptyValue(structFieldValue) || (fieldMap != nil && fieldMap[structField.Name]) {
			fieldValues[structField.Name] = structFieldValue.Interface()
		}
	})
	return
}

func ScanDefToTable(rv reflect.Value, table *Table) {
	ForEachStructFieldValue(reflect.Indirect(rv), func(structFieldValue reflect.Value, structField reflect.StructField, columnName string, tagValue string) {
		table.AddCol(Col(columnName).Field(structField.Name).Type(structFieldValue.Interface(), tagValue))
	})

	if rv.CanAddr() {
		addr := rv.Addr()
		if addr.CanInterface() {
			i := addr.Interface()

			if withComments, ok := i.(WithComments); ok {
				for fieldName, comment := range withComments.Comments() {
					field := table.F(fieldName)
					if field != nil {
						field.Comment = comment
					}
				}
			}

			if primaryKeyHook, ok := i.(WithPrimaryKey); ok {
				cols, err := table.Fields(primaryKeyHook.PrimaryKey()...)
				if err != nil {
					panic(fmt.Errorf("invalid primary key of table %s: %s", table.Name, err))
				}
				table.AddKey(PrimaryKey(cols))
			}

			if uniqueIndexesHook, ok := i.(WithUniqueIndexes); ok {
				for indexNameAndMethod, fieldNames := range uniqueIndexesHook.UniqueIndexes() {
					indexName, method := ResolveIndexNameAndMethod(indexNameAndMethod)
					cols, err := table.Fields(fieldNames...)
					if err != nil {
						panic(fmt.Errorf("invalid unique index %s of table %s: %s", indexName, table.Name, err))
					}
					table.AddKey(UniqueIndex(indexName, cols).Using(method))
				}
			}

			if indexesHook, ok := i.(WithIndexes); ok {
				for indexNameAndMethod, fieldNames := range indexesHook.Indexes() {
					indexName, method := ResolveIndexNameAndMethod(indexNameAndMethod)
					cols, err := table.Fields(fieldNames...)
					if err != nil {
						panic(fmt.Errorf("invalid index %s of table %s: %s", indexName, table.Name, err))
					}
					table.AddKey(Index(indexName, cols).Using(method))
				}
			}
		}
	}
}

func ResolveIndexNameAndMethod(n string) (name string, method string) {
	nameAndMethod := strings.Split(n, "/")
	name = nameAndMethod[0]
	if len(nameAndMethod) > 1 {
		method = nameAndMethod[1]
	}
	return
}
