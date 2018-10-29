package builder

import (
	"math"
)

type SqlAssignment interface {
	SqlExpr
	SqlAssignmentMarker
}

type SqlAssignmentMarker interface {
	asCondition()
}

func AsAssignment(ex SqlExpr) *Assignment {
	if ex.IsNil() {
		return &Assignment{SqlExpr: Expr("")}
	}
	return &Assignment{SqlExpr: ex}
}

type Assignment struct {
	SqlExpr
	SqlAssignmentMarker
}

func (a *Assignment) IsNil() bool {
	return a == nil || a.SqlExpr.IsNil()
}

type Assignments []*Assignment

func (assigns Assignments) IsNil() bool {
	return len(assigns) == 0
}

func (assigns Assignments) Expr() *Ex {
	e := Expr("")
	for i, assignment := range assigns {
		if i > 0 {
			e.WriteString(", ")
		}
		e.WriteExpr(assignment)
	}
	return e
}

func ColumnsAndValues(cols *Columns, values ...interface{}) *Assignment {
	if cols.IsNil() {
		return nil
	}

	n := cols.Len()

	groupCount := int(math.Round(float64(len(values)) / float64(n)))

	expr := Expr("")

	expr.WriteGroup(func(e *Ex) {
		expr.WriteExpr(cols)
	})

	expr.WriteString(" VALUES ")

	for i := 0; i < groupCount; i++ {
		if i > 0 {
			expr.WriteByte(',')
		}
		expr.WriteGroup(func(e *Ex) {
			for j := 0; j < n; j++ {
				e.WriteHolder(j)
			}
		})
	}

	expr.AppendArgs(values...)

	return AsAssignment(expr)
}
