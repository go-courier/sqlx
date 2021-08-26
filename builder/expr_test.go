package builder_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	. "github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/onsi/gomega"
)

func TestResolveExpr(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		gomega.NewWithT(t).Expect(ResolveExpr(nil)).To(gomega.BeNil())
	})
}

func TestEx(t *testing.T) {
	t.Run("empty query", func(t *testing.T) {
		gomega.NewWithT(t).Expect(
			Expr(""),
		).To(BeExpr(""))
	})

	t.Run("flatten slice", func(t *testing.T) {
		gomega.NewWithT(t).Expect(
			Expr(`#ID IN (?)`, []int{28, 29, 30}),
		).To(BeExpr("#ID IN (?,?,?)", 28, 29, 30))
	})

	t.Run("flatten should skip for bytes", func(t *testing.T) {
		gomega.NewWithT(t).Expect(
			Expr(`#ID = (?)`, []byte("")),
		).To(BeExpr("#ID = (?)", []byte("")))
	})

	t.Run("flatten with sub expr ", func(t *testing.T) {
		gomega.NewWithT(t).Expect(
			Expr(`#ID = ?`, Expr("#ID + ?", 1)),
		).To(BeExpr("#ID = #ID + ?", 1))
	})

	t.Run("flatten with ValuerExpr", func(t *testing.T) {
		gomega.NewWithT(t).Expect(
			Expr(`#Point = ?`, Point{1, 1}),
		).To(BeExpr("#Point = ST_GeomFromText(?)", Point{1, 1}))
	})
}

func BenchmarkEx(b *testing.B) {
	b.Run("empty query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Expr("").Ex(context.Background())
		}
	})

	b.Run("flatten slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Expr(`#ID IN (?)`, []int{28, 29, 30}).Ex(context.Background())
		}
	})

	b.Run("flatten with sub expr", func(b *testing.B) {
		b.Run("raw", func(b *testing.B) {
			eb := Expr("")
			eb.Grow(2)

			eb.WriteQuery("#ID > ?")
			eb.WriteQuery(" AND ")
			eb.WriteQuery("#ID < ?")

			eb.AppendArgs(1, 10)

			rawBuild := func() *Ex {
				return eb.Ex(context.Background())
			}

			clone := func(ex *Ex) *Ex {
				return Expr(ex.Query(), ex.Args()...).Ex(context.Background())
			}

			b.Run("clone", func(b *testing.B) {
				ex := rawBuild()

				for i := 0; i < b.N; i++ {
					_ = clone(ex)
				}
			})
		})

		b.Run("IsNilExpr", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				IsNilExpr(Expr(`#ID > ?`, 1))
			}
		})

		b.Run("by chain", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				e := AsCond(Expr(`#ID > ?`, 1)).And(AsCond(Expr(`#ID < ?`, 10)))
				e.Ex(context.Background())
			}
		})

		b.Run("by expr", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				e := And(
					Col("f_id").Lt(0),
					Col("f_id").In([]int{1, 2, 3}),
				)
				e.Ex(context.Background())
			}
		})

		b.Run("by expr without re created", func(b *testing.B) {
			left := Col("f_id").Lt(0)
			right := Col("f_id").In([]int{1, 2, 3})

			b.Run("single", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					left.Ex(context.Background())
				}
			})

			b.Run("composed", func(b *testing.B) {
				e := And(left, left, right, right)

				b.Log(e.Ex(context.Background()).Query())

				for i := 0; i < b.N; i++ {
					e.Ex(context.Background())
				}
			})
		})
	})
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
