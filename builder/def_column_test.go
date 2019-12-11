package builder_test

import (
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	. "github.com/go-courier/sqlx/v2/builder"
	"github.com/onsi/gomega"
)

func TestColumnTypeFromTypeAndTag(t *testing.T) {
	cases := map[string]*ColumnType{
		`,deprecated=f_target_env_id`: &ColumnType{
			Type:              reflect.TypeOf(1),
			DeprecatedActions: &DeprecatedActions{RenameTo: "f_target_env_id"},
		},
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
			gomega.NewWithT(t).Expect(ColumnTypeFromTypeAndTag(ct.Type, tagValue)).To(gomega.Equal(ct))
		})
	}
}
