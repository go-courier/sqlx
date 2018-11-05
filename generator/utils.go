package generator

import (
	"fmt"
	"go/types"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-courier/codegen"
	"github.com/go-courier/sqlx/v2/builder"
)

var (
	defRegexp = regexp.MustCompile(`@def ([^\n]+)`)
)

type Keys struct {
	Primary       []string
	Indexes       builder.Indexes
	UniqueIndexes builder.Indexes
}

func (ks *Keys) PatchUniqueIndexesWithSoftDelete(softDeleteField string) {
	if len(ks.UniqueIndexes) > 0 {
		for name, fieldNames := range ks.UniqueIndexes {
			ks.UniqueIndexes[name] = stringUniq(append(fieldNames, softDeleteField))
		}
	}
}

func (ks *Keys) Bind(table *builder.Table) {
	if len(ks.Primary) > 0 {
		cols, err := table.Fields(ks.Primary...)
		if err != nil {
			panic(fmt.Errorf("%s, please check primary def", err))
		}
		ks.Primary = cols.FieldNames()
		table.AddKey(builder.PrimaryKey(cols))
	}

	if len(ks.UniqueIndexes) > 0 {
		for indexNameAndMethod, fieldNames := range ks.UniqueIndexes {
			indexName, method := builder.ResolveIndexNameAndMethod(indexNameAndMethod)
			cols, err := table.Fields(fieldNames...)
			if err != nil {
				panic(fmt.Errorf("%s, please check unique_index def", err))
			}
			ks.UniqueIndexes[indexNameAndMethod] = cols.FieldNames()
			table.AddKey(builder.UniqueIndex(indexName, cols).Using(method))
		}
	}

	if len(ks.Indexes) > 0 {
		for indexNameAndMethod, fieldNames := range ks.Indexes {
			indexName, method := builder.ResolveIndexNameAndMethod(indexNameAndMethod)
			cols, err := table.Fields(fieldNames...)
			if err != nil {
				panic(fmt.Errorf("%s, please check index def", err))
			}
			ks.Indexes[indexNameAndMethod] = cols.FieldNames()
			table.AddKey(builder.Index(indexName, cols).Using(method))
		}
	}
}

func parseKeysFromDoc(doc string) *Keys {
	ks := &Keys{}
	matches := defRegexp.FindAllStringSubmatch(doc, -1)

	for _, subMatch := range matches {
		if len(subMatch) == 2 {
			defs := defSplit(subMatch[1])

			switch strings.ToLower(defs[0]) {
			case "primary":
				if len(defs) < 2 {
					panic(fmt.Errorf("primary at lease 1 Field"))
				}
				ks.Primary = defs[1:]
			case "unique_index":
				if len(defs) < 3 {
					panic(fmt.Errorf("unique indexes at lease 1 Field"))
				}
				if ks.UniqueIndexes == nil {
					ks.UniqueIndexes = builder.Indexes{}
				}
				ks.UniqueIndexes[defs[1]] = defs[2:]
			case "index":
				if len(defs) < 3 {
					panic(fmt.Errorf("index at lease 1 Field"))
				}
				if ks.Indexes == nil {
					ks.Indexes = builder.Indexes{}
				}
				ks.Indexes[defs[1]] = defs[2:]
			}
		}
	}
	return ks
}

func defSplit(def string) (defs []string) {
	vs := strings.Split(def, " ")
	for _, s := range vs {
		if s != "" {
			defs = append(defs, s)
		}
	}
	return
}

func toDefaultTableName(name string) string {
	return codegen.LowerSnakeCase("t_" + name)
}

func forEachStructField(structType *types.Struct, fn func(fieldVar *types.Var, columnName string, tagValue string)) {
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		tag := structType.Tag(i)
		if field.Exported() {
			structTag := reflect.StructTag(tag)
			tagValue, exists := structTag.Lookup("db")
			if exists {
				if tagValue != "-" {
					fn(field, builder.GetColumnName(field.Name(), tagValue), tagValue)
				}
			} else if field.Anonymous() {
				if nextStructType, ok := field.Type().Underlying().(*types.Struct); ok {
					forEachStructField(nextStructType, fn)
				}
				continue
			}
		}
	}
}

func stringPartition(list []string, checker func(item string, i int) bool) ([]string, []string) {
	newLeftList := make([]string, 0)
	newRightList := make([]string, 0)
	for i, item := range list {
		if checker(item, i) {
			newLeftList = append(newLeftList, item)
		} else {
			newRightList = append(newRightList, item)
		}
	}
	return newLeftList, newRightList
}

func stringFilter(list []string, checker func(item string, i int) bool) []string {
	newList, _ := stringPartition(list, checker)
	return newList
}

func stringUniq(list []string) (result []string) {
	strMap := make(map[string]bool)
	for _, str := range list {
		strMap[str] = true
	}

	for i := range list {
		str := list[i]
		if _, ok := strMap[str]; ok {
			delete(strMap, str)
			result = append(result, str)
		}
	}
	return
}
