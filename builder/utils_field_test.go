package builder

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-courier/x/ptr"
	typex "github.com/go-courier/x/types"
	g "github.com/onsi/gomega"
)

type SubSubSub struct {
	X string `db:"f_x"`
}

type SubSub struct {
	SubSubSub
}

type Sub struct {
	SubSub
	A string `db:"f_a"`
}

type PtrSub struct {
	B   []string          `db:"f_b"`
	Map map[string]string `db:"f_b_map"`
}

type P struct {
	Sub
	*PtrSub
	C *string `db:"f_c"`
}

var p *P

func init() {
	p = &P{}
	p.X = "x"
	p.A = "a"
	p.PtrSub = &PtrSub{
		Map: map[string]string{
			"1": "!",
		},
		B: []string{"b"},
	}
	p.C = ptr.String("c")
}

func TestTableFieldsFor(t *testing.T) {
	fields := StructFieldsFor(context.Background(), typex.FromRType(reflect.TypeOf(p)))

	rv := reflect.ValueOf(p)

	g.NewWithT(t).Expect(fields).To(g.HaveLen(5))

	g.NewWithT(t).Expect(fields[0].Name).To(g.Equal("f_x"))
	g.NewWithT(t).Expect(fields[0].FieldValue(rv).Interface()).To(g.Equal(p.X))

	g.NewWithT(t).Expect(fields[1].Name).To(g.Equal("f_a"))
	g.NewWithT(t).Expect(fields[1].FieldValue(rv).Interface()).To(g.Equal(p.A))

	g.NewWithT(t).Expect(fields[2].Name).To(g.Equal("f_b"))
	g.NewWithT(t).Expect(fields[2].FieldValue(rv).Interface()).To(g.Equal(p.B))

	g.NewWithT(t).Expect(fields[3].Name).To(g.Equal("f_b_map"))
	g.NewWithT(t).Expect(fields[3].FieldValue(rv).Interface()).To(g.Equal(p.Map))

	g.NewWithT(t).Expect(fields[4].Name).To(g.Equal("f_c"))
	g.NewWithT(t).Expect(fields[4].FieldValue(rv).Interface()).To(g.Equal(p.C))
}

func BenchmarkTableFieldsFor(b *testing.B) {
	typeP := reflect.TypeOf(p)

	_ = StructFieldsFor(context.Background(), typex.FromRType(typeP))

	//b.Log(typex.FromRType(reflect.TypeOf(p)).Unwrap() == typex.FromRType(reflect.TypeOf(p)).Unwrap())

	b.Run("StructFieldsFor", func(b *testing.B) {
		typP := typex.FromRType(typeP)

		for i := 0; i < b.N; i++ {
			_ = StructFieldsFor(context.Background(), typP)
		}
	})
}
