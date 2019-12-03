package builder

import (
	"testing"

	"github.com/go-courier/testingx"
	"github.com/onsi/gomega"
)

func TestColumns(t *testing.T) {
	columns := Columns{}

	t.Run("empty columns", testingx.It(func(t *testingx.T) {
		t.Expect(columns.Len()).To(gomega.Equal(0))
		t.Expect(columns.AutoIncrement()).To(gomega.BeNil())
	}))

	t.Run("added cols", testingx.It(func(t *testingx.T) {
		columns.Add(
			Col("F_id").Field("ID").Type(1, `,autoincrement`),
		)

		autoIncrementCol := columns.AutoIncrement()

		t.Expect(autoIncrementCol).NotTo(gomega.BeNil())
		t.Expect(autoIncrementCol.Name).To(gomega.Equal("f_id"))

		t.Run("get col by FieldName", testingx.It(func(t *testingx.T) {

			t.Expect(columns.F("ID2")).To(gomega.BeNil())

			t.Expect(MustCols(columns.Fields("ID2")).Len()).To(gomega.Equal(0))
			t.Expect(MustCols(columns.Fields()).Len()).To(gomega.Equal(1))

			t.Expect(MustCols(columns.Fields("ID2")).List()).To(gomega.HaveLen(0))
			t.Expect(MustCols(columns.Fields()).Len()).To(gomega.Equal(1))
		}))

		t.Run("get col by ColName", testingx.It(func(t *testingx.T) {
			t.Expect(MustCols(columns.Cols("F_id")).Len()).To(gomega.Equal(1))
			t.Expect(MustCols(columns.Cols()).Len()).To(gomega.Equal(1))
			t.Expect(MustCols(columns.Cols()).List()).To(gomega.HaveLen(1))

			t.Expect(MustCols(columns.Cols()).FieldNames()).To(gomega.Equal([]string{"ID"}))
		}))
	}))
}

func MustCols(cols *Columns, err error) *Columns {
	return cols
}
