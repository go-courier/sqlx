package builder

import (
	"context"
	"database/sql/driver"
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
		tagDB := f.Tags["db"]
		if tagDB != "" && tagDB != "-" {
			tfs = append(tfs, f)
		}
		return true
	})

	tf.cache.Store(underlying, tfs)

	return tfs
}

var typeModel = reflect.TypeOf((*Model)(nil)).Elem()
var driverValuer = reflect.TypeOf((*driver.Valuer)(nil)).Elem()

func EachStructField(ctx context.Context, tpe typesx.Type, each func(p *StructField) bool) {
	if tpe.Kind() != reflect.Struct {
		panic(fmt.Errorf("model %s must be a struct", tpe.Name()))
	}

	var walk func(tpe typesx.Type, modelLoc []int, parents ...int)

	walk = func(tpe typesx.Type, modelLoc []int, parents ...int) {
		if ok := tpe.Implements(typesx.FromRType(typeModel)); ok {
			modelLoc = parents
		}

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

			if (f.Anonymous() || f.Type().Name() == f.Name()) && (!hasDB) {
				fieldType := f.Type()

				if !fieldType.Implements(typesx.FromRType(driverValuer)) {
					for fieldType.Kind() == reflect.Ptr {
						fieldType = fieldType.Elem()
					}

					if fieldType.Kind() == reflect.Struct {
						walk(fieldType, modelLoc, loc...)
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

			p.Loc = make([]int, len(loc))
			copy(p.Loc, loc)

			p.ModelLoc = make([]int, len(modelLoc))
			copy(p.ModelLoc, modelLoc)

			p.ColumnType = *ColumnTypeFromTypeAndTag(p.Type, string(tagDB))

			if !each(p) {
				break
			}
		}
	}

	walk(tpe, []int{})
}

type StructField struct {
	Name       string
	FieldName  string
	Type       typesx.Type
	Field      typesx.StructField
	Tags       map[string]reflectx.StructTag
	Loc        []int
	ModelLoc   []int
	ColumnType ColumnType
}

func fieldValue(structReflectValue reflect.Value, locs []int) reflect.Value {
	n := len(locs)

	if n == 0 {
		return structReflectValue
	}

	if n < 0 {
		return reflect.Value{}
	}

	structReflectValue = reflectx.Indirect(structReflectValue)

	fv := structReflectValue

	for i := 0; i < n; i++ {
		loc := locs[i]
		fv = fv.Field(loc)

		// last loc should keep ptr value
		if i < n-1 {
			for fv.Kind() == reflect.Ptr {
				// notice the ptr struct ensure only for Ptr Anonymous StructField
				if fv.IsNil() {
					fv.Set(reflectx.New(fv.Type()))
				}
				fv = fv.Elem()
			}
		}
	}

	return fv
}

func (p *StructField) FieldValue(structReflectValue reflect.Value) reflect.Value {
	return fieldValue(structReflectValue, p.Loc)
}

func (p *StructField) FieldModelValue(structReflectValue reflect.Value) reflect.Value {
	return fieldValue(structReflectValue, p.ModelLoc)
}
