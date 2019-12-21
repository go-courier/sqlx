package builder

import (
	"context"
	"strconv"
)

type LimitAddition struct {
}

func (LimitAddition) AdditionType() AdditionType {
	return AdditionLimit
}

func Limit(rowCount int64) *limit {
	return &limit{rowCount: rowCount}
}

var _ Addition = (*limit)(nil)

type limit struct {
	LimitAddition

	// LIMIT
	rowCount int64
	// OFFSET
	offsetCount int64
}

func (l limit) Offset(offset int64) *limit {
	l.offsetCount = offset
	return &l
}

func (l *limit) IsNil() bool {
	return l == nil || l.rowCount <= 0
}

func (l *limit) Ex(ctx context.Context) *Ex {
	e := Expr("LIMIT ")

	e.WriteString(strconv.FormatInt(l.rowCount, 10))

	if l.offsetCount > 0 {
		e.WriteString(" OFFSET ")
		e.WriteString(strconv.FormatInt(l.offsetCount, 10))
	}

	return e.Ex(ctx)
}
