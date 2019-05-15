package builder

import (
	"sort"
	"strconv"
)

type Addition interface {
	SqlExpr
	weight() additionWeight
}

type additionWeight int

const (
	joinStmt additionWeight = iota
	whereStmt
	groupByStmt
	orderByStmt
	limitStmt
	onConflictStmt
	otherStmt
	commentStmt
)

type Additions []Addition

func (additions Additions) IsNil() bool {
	return len(additions) == 0
}

func (additions Additions) Expr() *Ex {
	finalAdditions := Additions{}
	for i := range additions {
		if additions[i].IsNil() {
			continue
		}
		finalAdditions = append(finalAdditions, additions[i])
	}
	sort.Sort(finalAdditions)

	if finalAdditions.IsNil() {
		return nil
	}

	ex := Expr("")
	for i := range finalAdditions {
		ex.WriteByte(' ')
		ex.WriteExpr(finalAdditions[i])
	}
	return ex
}

func (additions Additions) Len() int {
	return len(additions)
}

func (additions Additions) Less(i, j int) bool {
	return additions[i].weight() < additions[j].weight()
}

func (additions Additions) Swap(i, j int) {
	additions[i], additions[j] = additions[j], additions[i]
}

func Where(c SqlCondition) *where {
	return &where{
		SqlCondition: c,
	}
}

var _ Addition = (*where)(nil)

type where struct {
	SqlCondition
	WhereAddition
}

type WhereAddition struct{}

func (WhereAddition) weight() additionWeight {
	return whereStmt
}

func (w *where) IsNil() bool {
	return w == nil || w.SqlCondition.IsNil()
}

func (w *where) Expr() *Ex {
	if w.IsNil() {
		return nil
	}
	e := Expr("WHERE ")
	e.WriteExpr(w.SqlCondition)
	return e
}

func GroupBy(groups ...SqlExpr) *groupBy {
	return &groupBy{
		groups: groups,
	}
}

var _ Addition = (*groupBy)(nil)

type groupBy struct {
	GroupByAddition
	groups []SqlExpr
	// HAVING
	havingCond SqlCondition
}

type GroupByAddition struct {
}

func (GroupByAddition) weight() additionWeight {
	return groupByStmt
}

func (g groupBy) Having(cond *Condition) *groupBy {
	g.havingCond = cond
	return &g
}

func (g *groupBy) IsNil() bool {
	return g == nil || len(g.groups) == 0
}

func (g *groupBy) Expr() *Ex {
	if g.IsNil() {
		return nil
	}
	expr := Expr("GROUP BY ")

	for i, order := range g.groups {
		if i > 0 {
			expr.WriteByte(',')
		}
		expr.WriteExpr(order)
	}

	if !(g.havingCond == nil || g.havingCond.IsNil()) {
		expr.WriteString(" HAVING ")
		expr.WriteExpr(g.havingCond)
	}

	return expr
}

func OrderBy(orders ...*order) *orderBy {
	return &orderBy{
		orders: orders,
	}
}

var _ Addition = (*orderBy)(nil)

type orderBy struct {
	OrderByAddition
	orders []*order
}

type OrderByAddition struct {
}

func (OrderByAddition) weight() additionWeight {
	return orderByStmt
}

func (o *orderBy) IsNil() bool {
	return o == nil || len(o.orders) == 0
}

func (o *orderBy) Expr() *Ex {
	if o.IsNil() {
		return nil
	}
	expr := Expr("ORDER BY ")
	for i, order := range o.orders {
		if i > 0 {
			expr.WriteRune(',')
		}
		expr.WriteExpr(order)
	}
	return expr
	return nil
}

func AscOrder(expr SqlExpr) *order {
	return &order{
		expr: expr,
		typ:  "ASC",
	}
}

func DescOrder(expr SqlExpr) *order {
	return &order{
		expr: expr,
		typ:  "DESC",
	}
}

var _ SqlExpr = (*order)(nil)

type order struct {
	expr SqlExpr
	typ  string
}

func (o *order) IsNil() bool {
	return o == nil || o.expr.IsNil()
}

func (o *order) Expr() *Ex {
	if o.IsNil() {
		return nil
	}

	e := Expr("")

	e.WriteGroup(func(e *Ex) {
		e.WriteExpr(o.expr)
	})

	if o.typ != "" {
		e.WriteRune(' ')
		e.WriteString(o.typ)
	}
	return e
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

type LimitAddition struct {
}

func (LimitAddition) weight() additionWeight {
	return limitStmt
}

func (l limit) Offset(offset int64) *limit {
	l.offsetCount = offset
	return &l
}

func (l *limit) IsNil() bool {
	return l == nil || l.rowCount <= 0
}

func (l *limit) Expr() *Ex {
	if l.IsNil() {
		return nil
	}

	e := Expr("LIMIT ")
	e.WriteString(strconv.FormatInt(l.rowCount, 10))

	if l.offsetCount > 0 {
		e.WriteString(" OFFSET ")
		e.WriteString(strconv.FormatInt(l.offsetCount, 10))
	}
	return e
}

func OnConflict(columns *Columns) *onConflict {
	return &onConflict{
		columns: columns,
	}
}

type onConflict struct {
	OnConflictAddition
	columns     *Columns
	doNothing   bool
	assignments Assignments
}

func (o onConflict) DoNothing() *onConflict {
	o.doNothing = true
	return &o
}

func (o onConflict) DoUpdateSet(assignments ...*Assignment) *onConflict {
	o.assignments = Assignments(assignments)
	return &o
}

type OnConflictAddition struct {
}

func (OnConflictAddition) weight() additionWeight {
	return onConflictStmt
}

func (o *onConflict) IsNil() bool {
	return o == nil || o.columns == nil || (!o.doNothing && o.assignments.IsNil())
}

func (o *onConflict) Expr() *Ex {
	if o.IsNil() {
		return nil
	}

	e := Expr("ON CONFLICT ")
	e.WriteGroup(func(e *Ex) {
		e.WriteExpr(o.columns)
	})

	e.WriteString(" DO ")

	if o.doNothing {
		e.WriteString("NOTHING")
	} else {
		e.WriteString("UPDATE SET ")
		e.WriteExpr(o.assignments)
	}
	return e
}

func AsAddition(expr SqlExpr) *OtherAddition {
	return &OtherAddition{
		SqlExpr: expr,
	}
}

type OtherAddition struct {
	SqlExpr
}

func (OtherAddition) weight() additionWeight {
	return otherStmt
}

func (a *OtherAddition) IsNil() bool {
	return a == nil || a.SqlExpr.IsNil()
}

func Comment(c string) comment {
	return comment(c)
}

var _ Addition = comment("")

type comment string

func (comment) weight() additionWeight {
	return commentStmt
}

func (c comment) IsNil() bool {
	return len(c) == 0
}

func (c comment) Expr() *Ex {
	if c.IsNil() {
		return nil
	}
	e := Expr("")
	e.WhiteComments(string(c))
	return e
}
