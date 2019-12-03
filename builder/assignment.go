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

func AsAssignment(expr SqlExpr) *Assignment {
	if IsNilExpr(expr) {
		return &Assignment{SqlExpr: Expr("")}
	}
	return &Assignment{SqlExpr: expr}
}

type Assignments []*Assignment

type Assignment struct {
	SqlAssignmentMarker
	SqlExpr
}

func (a *Assignment) IsNil() bool {
	return a == nil || IsNilExpr(a.SqlExpr)
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
