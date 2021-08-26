package builder

import (
	"context"
	"fmt"
	"go/ast"
	"reflect"
	"strings"
	"sync"

	reflectx "github.com/go-courier/x/reflect"
	typesx "github.com/go-courier/x/types"
)

func StructFieldsFor(ctx context.Context, typ typesx.Type) []*StructField {
	return defaultStructFieldsFactory.TableFieldsFor(ctx, typ)
}

var defaultStructFieldsFactory = &StructFieldsFactory{}

type StructFieldsFactory struct {
	cache sync.Map
}

func (tf *StructFieldsFactory) TableFieldsFor(ctx context.Context, typ typesx.Type) []*StructField {
	typ = typesx.Deref(typ)

	underlying := typ.Unwrap()

	if v, ok := tf.cache.Load(underlying); ok {
		return v.([]*StructField)
	}

	tfs := make([]*StructField, 0)

	EachStructField(ctx, typ, func(f *StructField) bool {
		tfs = append(tfs, f)
		return true
	})

	tf.cache.Store(underlying, tfs)

	return tfs
}

func EachStructField(ctx context.Context, tpe typesx.Type, each func(p *StructField) bool) {
	if tpe.Kind() != reflect.Struct {
		panic(fmt.Errorf("model %s must be a struct", tpe.Name()))
	}

	var walk func(tpe typesx.Type, parents ...int)

	walk = func(tpe typesx.Type, parents ...int) {
		for i := 0; i < tpe.NumField(); i++ {
			f := tpe.Field(i)

			if !ast.IsExported(f.Name()) {
				continue
			}

			loc := append(parents, i)

			tags := reflectx.ParseStructTags(string(f.Tag()))

			displayName := f.Name()

			tagDB, hasDB := tags["db"]
			if hasDB {
				if name := tagDB.Name(); name == "-" {
					// skip name:"-"
					continue
				} else {
					if name != "" {
						displayName = name
					}
				}
			}

			if f.Anonymous() && (!hasDB) {
				fieldType := f.Type()

				_, ok := typesx.EncodingTextMarshalerTypeReplacer(fieldType)

				if !ok {
					for fieldType.Kind() == reflect.Ptr {
						fieldType = fieldType.Elem()
					}

					if fieldType.Kind() == reflect.Struct {
						walk(fieldType, loc...)
						continue
					}
				}
			}

			p := &StructField{}
			p.FieldName = f.Name()
			p.Type = f.Type()
			p.Field = f
			p.Tags = tags
			p.Name = strings.ToLower(displayName)
			p.Loc = loc
			p.ColumnType = *ColumnTypeFromTypeAndTag(p.Type, string(tagDB))

			if !each(p) {
				break
			}
		}
	}

	walk(tpe)
}

type StructField struct {
	Name       string
	FieldName  string
	Type       typesx.Type
	Field      typesx.StructField
	Tags       map[string]reflectx.StructTag
	Loc        []int
	ColumnType ColumnType
}

func (p *StructField) FieldValue(structReflectValue reflect.Value) reflect.Value {
	structReflectValue = reflectx.Indirect(structReflectValue)

	n := len(p.Loc)

	fieldValue := structReflectValue

	for i := 0; i < n; i++ {
		loc := p.Loc[i]
		fieldValue = fieldValue.Field(loc)

		// last loc should keep ptr value
		if i < n-1 {
			for fieldValue.Kind() == reflect.Ptr {
				// notice the ptr struct ensure only for Ptr Anonymous StructField
				if fieldValue.IsNil() {
					fieldValue.Set(reflectx.New(fieldValue.Type()))
				}
				fieldValue = fieldValue.Elem()
			}
		}
	}

	return fieldValue
}
