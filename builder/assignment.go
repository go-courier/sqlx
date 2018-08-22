package builder

import (
	"bytes"
	"math"
)

type Assignment Expression

func (a Assignment) Expr() *Expression {
	return (*Expression)(&a)
}

type Assignments []*Assignment

func (assigns Assignments) Expr() (e *Expression) {
	e = Expr("")
	for i, assignment := range assigns {
		joiner := ", "
		if i == 0 {
			joiner = ""
		}
		e = MustJoinExpr(joiner, e, assignment)
	}
	return e
}

func ColumnsAndValues(cols Columns, values ...interface{}) *Assignment {
	n := cols.Len()
	holderGroup := HolderRepeat(n)

	groupCount := int(math.Round(float64(len(values)) / float64(n)))

	buf := bytes.NewBuffer(nil)
	for i := 0; i < groupCount; i++ {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('(')
		buf.WriteString(holderGroup)
		buf.WriteByte(')')
	}

	expr := MustJoinExpr(" VALUES ", cols.Group(), Expr(buf.String(), values...))
	return (*Assignment)(expr)
}
