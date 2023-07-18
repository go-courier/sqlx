package generator

import (
	"go/types"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-courier/codegen"
	"github.com/go-courier/sqlx/v2/builder"
)

var (
	defRegexp = regexp.MustCompile(`@def ([^\n]+)`)
	relRegexp = regexp.MustCompile(`@rel ([^\n]+)`)
)

type Keys struct {
	Primary       []string
	Indexes       builder.Indexes
	UniqueIndexes builder.Indexes
	Partition     []string
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
		key := &builder.Key{
			Name:     "primary",
			IsUnique: true,
			Def:      *builder.ParseIndexDef(ks.Primary...),
		}

		_ = key.Def.TableExpr(table)

		table.AddKey(key)
	}

	if len(ks.UniqueIndexes) > 0 {
		for indexNameAndMethod, fieldNames := range ks.UniqueIndexes {
			indexName, method := builder.ResolveIndexNameAndMethod(indexNameAndMethod)

			key := &builder.Key{
				Name:     indexName,
				Method:   method,
				IsUnique: true,
				Def:      *builder.ParseIndexDef(fieldNames...),
			}

			_ = key.Def.TableExpr(table)

			table.AddKey(key)
		}
	}

	if len(ks.Indexes) > 0 {
		for indexNameAndMethod, fieldNames := range ks.Indexes {
			indexName, method := builder.ResolveIndexNameAndMethod(indexNameAndMethod)

			key := &builder.Key{
				Name:   indexName,
				Method: method,
				Def:    *builder.ParseIndexDef(fieldNames...),
			}

			_ = key.Def.TableExpr(table)

			table.AddKey(key)
		}
	}
}

func parseColRelFromComment(doc string) (string, []string) {
	others := make([]string, 0)

	rel := ""

	for _, line := range strings.Split(doc, "\n") {
		if len(line) == 0 {
			continue
		}

		matches := relRegexp.FindAllStringSubmatch(line, 1)

		if matches == nil {
			others = append(others, line)
			continue
		}

		if len(matches) == 1 {
			rel = matches[0][1]
		}
	}

	return rel, others
}

func parseKeysFromDoc(doc string) (*Keys, []string) {
	ks := &Keys{}

	others := make([]string, 0)

	for _, line := range strings.Split(doc, "\n") {
		if len(line) == 0 {
			continue
		}

		matches := defRegexp.FindAllStringSubmatch(line, -1)

		if matches == nil {
			others = append(others, line)
			continue
		}

		for _, subMatch := range matches {
			if len(subMatch) == 2 {
				def := builder.ParseIndexDefine(subMatch[1])

				switch def.Kind {
				case "primary":
					ks.Primary = def.ToDefs()
				case "unique_index":
					if ks.UniqueIndexes == nil {
						ks.UniqueIndexes = builder.Indexes{}
					}
					ks.UniqueIndexes[def.ID()] = def.ToDefs()
				case "index":
					if ks.Indexes == nil {
						ks.Indexes = builder.Indexes{}
					}
					ks.Indexes[def.ID()] = def.ToDefs()
				case "partition":
					ks.Partition = append([]string{def.Name}, def.ToDefs()...)
				}
			}
		}
	}

	return ks, others
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
