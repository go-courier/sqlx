package builder

import (
	"testing"

	"github.com/onsi/gomega"
)

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
