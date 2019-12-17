package builder

import (
	"context"
	"math"
)

type SqlAssignment interface {
	SqlExpr
	SqlAssignmentMarker
}

type SqlAssignmentMarker interface {
	asCondition()
}

func WriteAssignments(e *Ex, assignments ...*Assignment) {
	count := 0

	for i := range assignments {
		a := assignments[i]

		if IsNilExpr(a) {
			continue
		}

		if count > 0 {
			e.WriteString(", ")
		}

		e.WriteExpr(a)
		count++
	}
}

func ColumnsAndValues(columnOrColumns SqlExpr, values ...interface{}) *Assignment {
	lenOfColumn := 1
	if canLen, ok := columnOrColumns.(interface{ Len() int }); ok {
		lenOfColumn = canLen.Len()
	}
	return &Assignment{columnOrColumns: columnOrColumns, lenOfColumn: lenOfColumn, values: values}
}

type Assignments []*Assignment

type Assignment struct {
	SqlAssignmentMarker

	columnOrColumns SqlExpr
	lenOfColumn     int

	values []interface{}
}

func (a *Assignment) IsNil() bool {
	return a == nil || IsNilExpr(a.columnOrColumns) || len(a.values) == 0
}

func (a *Assignment) Ex(ctx context.Context) *Ex {
	e := Expr("")

	useValues := TogglesFromContext(ctx).Is(ToggleUseValues)

	if useValues {
		e.WriteGroup(func(e *Ex) {
			e.WriteExpr(ExprBy(func(ctx context.Context) *Ex {
				return a.columnOrColumns.Ex(ContextWithToggles(ctx, Toggles{
					ToggleMultiTable: false,
				}))
			}))
		})

		if len(a.values) == 1 {
			if s, ok := a.values[0].(SelectStatement); ok {
				e.WriteByte(' ')
				e.WriteExpr(s)
				return e.Ex(ctx)
			}
		}

		e.WriteString(" VALUES ")

		groupCount := int(math.Round(float64(len(a.values)) / float64(a.lenOfColumn)))

		for i := 0; i < groupCount; i++ {
			if i > 0 {
				e.WriteByte(',')
			}

			e.WriteGroup(func(e *Ex) {
				for j := 0; j < a.lenOfColumn; j++ {
					e.WriteHolder(j)
				}
			})
		}

		e.AppendArgs(a.values...)

		return e.Ex(ctx)
	}

	e.WriteExpr(ExprBy(func(ctx context.Context) *Ex {
		return a.columnOrColumns.Ex(ContextWithToggles(ctx, Toggles{
			ToggleMultiTable: false,
		}))
	}))

	e.WriteString(" = ?")
	e.AppendArgs(a.values[0])

	return e.Ex(ctx)
}
