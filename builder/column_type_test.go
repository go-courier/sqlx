package builder

import (
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/stretchr/testify/require"
)

func TestColumnTypeFromTypeAndTag(t *testing.T) {
	cases := map[string]*ColumnType{
		`,autoincrement`: &ColumnType{
			Type:          reflect.TypeOf(1),
			AutoIncrement: true,
		},
		`,null`: &ColumnType{
			Type: reflect.TypeOf(float64(1.1)),
			Null: true,
		},
		`,size=2`: &ColumnType{
			Type:   reflect.TypeOf(""),
			Length: 2,
		},
		`,decimal=1`: &ColumnType{
			Type:    reflect.TypeOf(float64(1.1)),
			Decimal: 1,
		},
		`,default='1'`: &ColumnType{
			Type:    reflect.TypeOf(""),
			Default: ptr.String(`'1'`),
		},
	}

	for tagValue, ct := range cases {
		t.Run(tagValue, func(t *testing.T) {
			require.Equal(t, ColumnTypeFromTypeAndTag(ct.Type, tagValue), ct)
		})
	}
}
