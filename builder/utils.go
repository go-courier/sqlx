package builder

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"github.com/go-courier/reflectx"
)

type FieldValues map[string]interface{}

func ForEachStructFieldValue(rv reflect.Value, fn func(structFieldValue reflect.Value, structField reflect.StructField, columnName string, tagValue string)) {
	rv = reflect.Indirect(rv)
	structType := rv.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		if field.Type.Kind() == reflect.Interface {
			continue
		}

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
	return strings.ToLower(columnName)
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

func TableFromModel(model Model) *Table {
	tpe := reflect.TypeOf(model)
	if tpe.Kind() != reflect.Ptr {
		panic(fmt.Errorf("model %s must be a pointer", tpe.Name()))
	}
	tpe = tpe.Elem()
	if tpe.Kind() != reflect.Struct {
		panic(fmt.Errorf("model %s must be a struct", tpe.Name()))
	}

	table := T(model.TableName())
	table.Model = model

	ScanDefToTable(reflect.Indirect(reflect.ValueOf(model)), table)

	return table
}

func ScanDefToTable(rv reflect.Value, table *Table) {
	table.ModelName = rv.Type().Name()

	ForEachStructFieldValue(reflect.Indirect(rv), func(structFieldValue reflect.Value, structField reflect.StructField, columnName string, tagValue string) {
		table.AddCol(Col(columnName).Field(structField.Name).Type(structFieldValue.Interface(), tagValue))
	})

	if rv.CanAddr() {
		addr := rv.Addr()
		if addr.CanInterface() {
			i := addr.Interface()

			if withTableDescription, ok := i.(WithTableDescription); ok {
				desc := withTableDescription.TableDescription()
				table.Description = desc
			}

			if withComments, ok := i.(WithComments); ok {
				for fieldName, comment := range withComments.Comments() {
					field := table.F(fieldName)
					if field != nil {
						field.Comment = comment
					}
				}
			}

			if withColDescriptions, ok := i.(WithColDescriptions); ok {
				for fieldName, desc := range withColDescriptions.ColDescriptions() {
					field := table.F(fieldName)
					if field != nil {
						field.Description = desc
					}
				}
			}

			if withRelations, ok := i.(WithRelations); ok {
				for fieldName, rel := range withRelations.ColRelations() {
					field := table.F(fieldName)
					if field != nil {
						field.Relation = rel
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
