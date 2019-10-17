package builder

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-courier/reflectx"
)

func ColumnTypeFromTypeAndTag(typ reflect.Type, nameAndFlags string) *ColumnType {
	ct := &ColumnType{}
	ct.Type = reflectx.Deref(typ)

	v := reflect.New(ct.Type).Interface()

	if dataTypeDescriber, ok := v.(DataTypeDescriber); ok {
		ct.GetDataType = dataTypeDescriber.DataType
	}

	if strings.Index(nameAndFlags, ",") > -1 {
		for _, flag := range strings.Split(nameAndFlags, ",")[1:] {
			nameAndValue := strings.Split(flag, "=")
			switch strings.ToLower(nameAndValue[0]) {
			case "null":
				ct.Null = true
			case "autoincrement":
				ct.AutoIncrement = true
			case "deprecated":
				rename := ""
				if len(nameAndValue) > 1 {
					rename = nameAndValue[1]
				}
				ct.DeprecatedActions = &DeprecatedActions{RenameTo: rename}
			case "size":
				if len(nameAndValue) == 1 {
					panic(fmt.Errorf("missing size value"))
				}
				length, err := strconv.ParseUint(nameAndValue[1], 10, 64)
				if err != nil {
					panic(fmt.Errorf("invalid size value: %s", err))
				}
				ct.Length = length
			case "decimal":
				if len(nameAndValue) == 1 {
					panic(fmt.Errorf("missing size value"))
				}
				decimal, err := strconv.ParseUint(nameAndValue[1], 10, 64)
				if err != nil {
					panic(fmt.Errorf("invalid decimal value: %s", err))
				}
				ct.Decimal = decimal
			case "default":
				if len(nameAndValue) == 1 {
					panic(fmt.Errorf("missing default value"))
				}
				ct.Default = &nameAndValue[1]
			}
		}
	}

	return ct
}

type ColumnType struct {
	Type        reflect.Type
	GetDataType func(engine string) string

	Length  uint64
	Decimal uint64

	Default *string

	Null          bool
	AutoIncrement bool

	Comment string

	DeprecatedActions *DeprecatedActions
}

type DeprecatedActions struct {
	RenameTo string `name:"rename"`
}
