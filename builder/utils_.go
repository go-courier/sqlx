package builder

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	typesx "github.com/go-courier/x/types"

	contextx "github.com/go-courier/x/context"
	reflectx "github.com/go-courier/x/reflect"
)

type FieldValues map[string]interface{}

type StructFieldValue struct {
	Field     StructField
	TableName string
	Value     reflect.Value
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
	ctx := context.Background()

	fields := StructFieldsFor(ctx, typesx.FromRType(reflect.TypeOf(v)))

	e := Expr("")
	e.Grow(len(fields))

	i := 0

	ForEachStructFieldValue(context.Background(), reflect.ValueOf(v), func(field *StructFieldValue) {
		if i > 0 {
			e.WriteQuery(", ")
		}

		if field.TableName != "" {
			e.WriteQuery(field.TableName)
			e.WriteQueryByte('.')
			e.WriteQuery(field.Field.Name)
			e.WriteQuery(" AS ")
			e.WriteQuery(field.TableName)
			e.WriteQuery("__")
			e.WriteQuery(field.Field.Name)
		} else {
			e.WriteQuery(field.Field.Name)
		}

		i++
	})

	return e
}

func ForEachStructFieldValue(ctx context.Context, v interface{}, fn func(*StructFieldValue)) {
	rv, ok := v.(reflect.Value)
	if ok {
		if rv.Kind() == reflect.Ptr && rv.IsNil() {
			rv.Set(reflectx.New(rv.Type()))
		}
		v = rv.Interface()
	}

	if m, ok := v.(Model); ok {
		ctx = WithTableName(m.TableName())(ctx)
	}

	fields := StructFieldsFor(ctx, typesx.FromRType(reflect.TypeOf(v)))

	rv = reflectx.Indirect(reflect.ValueOf(v))

	for i := range fields {
		f := fields[i]

		tagDB := f.Tags["db"]

		if tagDB.HasFlag("deprecated") {
			continue
		}

		if tableAlias, ok := f.Tags["alias"]; ok {
			ctx = WithTableAlias(tableAlias.Name())(ctx)
		}

		sf := &StructFieldValue{}

		sf.Field = *f
		sf.Value = f.FieldValue(rv)
		sf.TableName = TableNameFromContext(ctx)

		if tableAlias := TableAliasFromContext(ctx); tableAlias != "" {
			sf.TableName = tableAlias
		}

		fn(sf)
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
	ForEachStructFieldValue(context.Background(), rv, func(sf *StructFieldValue) {
		if fieldMap != nil && fieldMap[sf.Field.FieldName] {
			fieldValues[sf.Field.FieldName] = sf.Value.Interface()
		}
	})
	return fieldValues
}

func FieldValuesFromStructByNonZero(structValue interface{}, excludes ...string) (fieldValues FieldValues) {
	fieldValues = FieldValues{}
	rv := reflect.Indirect(reflect.ValueOf(structValue))
	fieldMap := ToMap(excludes)
	ForEachStructFieldValue(context.Background(), rv, func(sf *StructFieldValue) {
		if !reflectx.IsEmptyValue(sf.Value) || (fieldMap != nil && fieldMap[sf.Field.FieldName]) {
			fieldValues[sf.Field.FieldName] = sf.Value.Interface()
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

	ScanDefToTable(table, model)

	return table
}

func ScanDefToTable(table *Table, i interface{}) {
	tpe := typesx.Deref(typesx.FromRType(reflect.TypeOf(i)))

	EachStructField(context.Background(), tpe, func(f *StructField) bool {
		table.AddCol(&Column{
			FieldName:  f.FieldName,
			Name:       f.Name,
			ColumnType: &f.ColumnType,
		})
		return true
	})

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
		table.AddKey(&Key{
			Name:     "primary",
			IsUnique: true,
			Def:      *ParseIndexDef(primaryKeyHook.PrimaryKey()...),
		})
	}

	if uniqueIndexesHook, ok := i.(WithUniqueIndexes); ok {
		for indexNameAndMethod, fieldNames := range uniqueIndexesHook.UniqueIndexes() {
			indexName, method := ResolveIndexNameAndMethod(indexNameAndMethod)

			table.AddKey(&Key{
				Name:     indexName,
				Method:   method,
				IsUnique: true,
				Def:      *ParseIndexDef(fieldNames...),
			})
		}
	}

	if indexesHook, ok := i.(WithIndexes); ok {
		for indexNameAndMethod, fieldNames := range indexesHook.Indexes() {
			indexName, method := ResolveIndexNameAndMethod(indexNameAndMethod)

			table.AddKey(&Key{
				Name:   indexName,
				Method: method,
				Def:    *ParseIndexDef(fieldNames...),
			})
		}
	}
}

func ResolveIndexNameAndMethod(n string) (name string, method string) {
	nameAndMethod := strings.Split(n, "/")
	name = strings.ToLower(nameAndMethod[0])
	if len(nameAndMethod) > 1 {
		method = nameAndMethod[1]
	}
	return
}

// ParseIndexDefine
// @def index i_xxx/BTREE Name
// @def index i_xxx USING GIST (#TEST gist_trgm_ops)
func ParseIndexDefine(def string) *IndexDefine {
	d := IndexDefine{}

	for i := strings.Index(def, " "); i != -1; i = strings.Index(def, " ") {
		part := def[0:i]

		if part != "" {
			if d.Kind == "" {
				d.Kind = part
			} else if d.Name == "" && d.Kind != "primary" {
				d.Name, d.Method = ResolveIndexNameAndMethod(part)
			} else {
				break
			}
		}

		def = def[i+1:]
	}

	d.IndexDef = *ParseIndexDef(strings.TrimSpace(def))

	return &d
}

type IndexDefine struct {
	Kind   string
	Name   string
	Method string
	IndexDef
}

func (i IndexDefine) ID() string {
	if i.Method != "" {
		return i.Name + "/" + i.Method
	}
	return i.Name
}
