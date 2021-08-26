package builder

import (
	"testing"

	"github.com/onsi/gomega"
)

func BenchmarkCols(b *testing.B) {
	columns := Columns{}

	columns.Add(
		Col("f_id").Field("ID").Type(1, `,autoincrement`),
		Col("f_name").Field("Name").Type(1, ``),
		Col("f_f1").Field("F1").Type(1, ``),
		Col("f_f2").Field("F2").Type(1, ``),
		Col("f_f3").Field("F3").Type(1, ``),
		Col("f_f4").Field("F4").Type(1, ``),
		Col("f_f5").Field("F5").Type(1, ``),
		Col("f_f6").Field("F6").Type(1, ``),
		Col("f_f7").Field("F7").Type(1, ``),
		Col("f_f8").Field("F8").Type(1, ``),
		Col("f_f9").Field("F9").Type(1, ``),
	)

	b.Run("pick", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = columns.F("F3")
		}
	})

	b.Run("multi pick", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = columns.Fields("ID", "Name")
		}
	})

	b.Run("multi pick all", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = columns.Fields()
		}
	})

}

func TestColumns(t *testing.T) {
	columns := Columns{}

	t.Run("empty columns", func(t *testing.T) {
		gomega.NewWithT(t).Expect(columns.Len()).To(gomega.Equal(0))
		gomega.NewWithT(t).Expect(columns.AutoIncrement()).To(gomega.BeNil())
	})

	t.Run("added cols", func(t *testing.T) {
		columns.Add(
			Col("F_id").Field("ID").Type(1, `,autoincrement`),
		)

		autoIncrementCol := columns.AutoIncrement()

		gomega.NewWithT(t).Expect(autoIncrementCol).NotTo(gomega.BeNil())
		gomega.NewWithT(t).Expect(autoIncrementCol.Name).To(gomega.Equal("f_id"))

		t.Run("get col by FieldName", func(t *testing.T) {

			gomega.NewWithT(t).Expect(columns.F("ID2")).To(gomega.BeNil())

			gomega.NewWithT(t).Expect(MustCols(columns.Fields("ID2")).Len()).To(gomega.Equal(0))
			gomega.NewWithT(t).Expect(MustCols(columns.Fields()).Len()).To(gomega.Equal(1))

			gomega.NewWithT(t).Expect(MustCols(columns.Fields("ID2")).List()).To(gomega.HaveLen(0))
			gomega.NewWithT(t).Expect(MustCols(columns.Fields()).Len()).To(gomega.Equal(1))
		})
		t.Run("get col by ColName", func(t *testing.T) {
			gomega.NewWithT(t).Expect(MustCols(columns.Cols("F_id")).Len()).To(gomega.Equal(1))
			gomega.NewWithT(t).Expect(MustCols(columns.Cols()).Len()).To(gomega.Equal(1))
			gomega.NewWithT(t).Expect(MustCols(columns.Cols()).List()).To(gomega.HaveLen(1))

			gomega.NewWithT(t).Expect(MustCols(columns.Cols()).FieldNames()).To(gomega.Equal([]string{"ID"}))
		})
	})
}

func MustCols(cols *Columns, err error) *Columns {
	return cols
}
