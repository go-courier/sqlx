package builder_test

import (
	"database/sql/driver"
	"fmt"
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
	"github.com/onsi/gomega"
)

func TestResolveExpr(t *testing.T) {
	t.Run("empty", testingx.It(func(t *testingx.T) {
		t.Expect(ResolveExpr(nil)).To(gomega.BeNil())
	}))
}

func TestEx(t *testing.T) {
	t.Run("empty query", testingx.It(func(t *testingx.T) {
		t.Expect(
			Expr(""),
		).To(BeExpr(""))
	}))

	t.Run("flatten slice", testingx.It(func(t *testingx.T) {
		t.Expect(
			Expr(`#ID IN (?)`, []int{28, 29, 30}),
		).To(BeExpr("#ID IN (?,?,?)", 28, 29, 30))
	}))

	t.Run("flatten should skip for bytes", testingx.It(func(t *testingx.T) {
		t.Expect(
			Expr(`#ID = (?)`, []byte("")),
		).To(BeExpr("#ID = (?)", []byte("")))
	}))

	t.Run("flatten with sub expr ", testingx.It(func(t *testingx.T) {
		t.Expect(
			Expr(`#ID = ?`, Expr("#ID + ?", 1)),
		).To(BeExpr("#ID = #ID + ?", 1))
	}))

	t.Run("flatten with ValuerExpr", testingx.It(func(t *testingx.T) {
		t.Expect(
			Expr(`#Point = ?`, Point{1, 1}),
		).To(BeExpr("#Point = ST_GeomFromText(?)", Point{1, 1}))
	}))
}

type Point struct {
	X float64
	Y float64
}

func (Point) DataType(engine string) string {
	return "POINT"
}

func (Point) ValueEx() string {
	return `ST_GeomFromText(?)`
}

func (p Point) Value() (driver.Value, error) {
	return fmt.Sprintf("POINT(%v %v)", p.X, p.Y), nil
}
