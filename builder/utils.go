package builder

import (
	"context"
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	contextx "github.com/go-courier/x/context"

	reflectx "github.com/go-courier/x/reflect"
)

type FieldValues map[string]interface{}

type StructField struct {
	Value      reflect.Value
	Field      reflect.StructField
	TableName  string
	ColumnName string
	TagValue   string
}

type contextKeyTableName struct{}

func WithTableName(tableName string) func(ctx context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return contextx.WithValue(ctx, contextKeyTableName{}, tableName)
	}
}

func TableNameFromContext(ctx context.Context) string {
	if tableName, ok := ctx.Value(contextKeyTableName{}).(string); ok {
		return tableName
	}
	return ""
}

type contextKeyTableAlias int

func WithTableAlias(tableName string) func(ctx context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return contextx.WithValue(ctx, contextKeyTableAlias(1), tableName)
	}
}

func TableAliasFromContext(ctx context.Context) string {
	if tableName, ok := ctx.Value(contextKeyTableAlias(1)).(string); ok {
		return tableName
	}
	return ""
}

func ColumnsByStruct(v interface{}) *Ex {
	e := Expr("")

	i := 0

	ForEachStructFieldValue(context.Background(), reflect.ValueOf(v), func(field *StructField) {
		if i > 0 {
			e.WriteString(", ")
		}

		if field.TableName != "" {
			_, _ = e.WriteString(field.TableName)
			_ = e.WriteByte('.')
			_, _ = e.WriteString(field.ColumnName)
			_, _ = e.WriteString(" AS ")
			_, _ = e.WriteString(field.TableName)
			_, _ = e.WriteString("__")
			_, _ = e.WriteString(field.ColumnName)
		} else {
			_, _ = e.WriteString(field.ColumnName)
		}

		i++
	})

	return e
}

func ForEachStructFieldValue(ctx context.Context, rv reflect.Value, fn func(*StructField)) {
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}

	rv = reflect.Indirect(rv)

	structType := rv.Type()

	if structType.Kind() == reflect.Struct {
		if m, ok := rv.Interface().(Model); ok {
			ctx = WithTableName(m.TableName())(ctx)
		}

		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)

			if field.Type.Kind() == reflect.Interface {
				continue
			}

			if tableAlias, ok := field.Tag.Lookup("alias"); ok {
				ctx = WithTableAlias(tableAlias)(ctx)
			}

			if ast.IsExported(field.Name) {
				fieldValue := rv.Field(i)

				tagValue, exists := field.Tag.Lookup("db")
				if exists {
					if tagValue != "-" && !strings.Contains(tagValue, ",deprecated") {
						sf := &StructField{}
						sf.Value = fieldValue
						sf.Field = field
						sf.TableName = TableNameFromContext(ctx)

						if tableAlias := TableAliasFromContext(ctx); tableAlias != "" {
							sf.TableName = tableAlias
						}

						sf.ColumnName = GetColumnName(field.Name, tagValue)
						sf.TagValue = tagValue

						fn(sf)
					}
				} else if field.Anonymous {
					ForEachStructFieldValue(ctx, fieldValue, fn)
				} else {
					if _, ok := fieldValue.Interface().(Model); ok {
						ForEachStructFieldValue(ctx, fieldValue, fn)
					}
				}
			}
		}
	}
}

func GetColumnName(fieldName, tagValue string) string {
	i := strings.Index(tagValue, ",")
	if tagValue != "" && (i > 0 || i == -1) {
		if i == -1 {
			return strings.ToLower(tagValue)
		}
		return strings.ToLower(tagValue[0:i])
	}
	return "f_" + strings.ToLower(fieldName)
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
	ForEachStructFieldValue(context.Background(), rv, func(sf *StructField) {
		if fieldMap != nil && fieldMap[sf.Field.Name] {
			fieldValues[sf.Field.Name] = sf.Value.Interface()
		}
	})
	return fieldValues
}

func FieldValuesFromStructByNonZero(structValue interface{}, excludes ...string) (fieldValues FieldValues) {
	fieldValues = FieldValues{}
	rv := reflect.Indirect(reflect.ValueOf(structValue))
	fieldMap := ToMap(excludes)
	ForEachStructFieldValue(context.Background(), rv, func(sf *StructField) {
		if !reflectx.IsEmptyValue(sf.Value) || (fieldMap != nil && fieldMap[sf.Field.Name]) {
			fieldValues[sf.Field.Name] = sf.Value.Interface()
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

	ForEachStructFieldValue(context.Background(), reflect.Indirect(rv), func(sf *StructField) {
		table.AddCol(Col(sf.ColumnName).Field(sf.Field.Name).Type(sf.Value.Interface(), sf.TagValue))
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
