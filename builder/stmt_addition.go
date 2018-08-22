package builder

import (
	"fmt"
	"sort"
)

const (
	ALL      = "ALL"
	DISTINCT = "DISTINCT"

	DISTINCTROW = "DISTINCTROW"

	LOW_PRIORITY  = "LOW_PRIORITY"
	DELAYED       = "LOW_PRIORITY"
	HIGH_PRIORITY = "HIGH_PRIORITY"

	QUICK = "QUICK"

	STRAIGHT_JOIN = "STRAIGHT_JOIN"

	SQL_SMALL_RESULT  = "SQL_SMALL_RESULT"
	SQL_BIG_RESULT    = "SQL_BIG_RESULT"
	SQL_BUFFER_RESULT = "SQL_BUFFER_RESULT"

	IGNORE = "IGNORE"
)

type Addition interface {
	SqlExpr
	weight() additionWeight
}

type Additions []Addition

func (a Additions) Expr() *Expression {
	sort.Sort(a)
	exprList := make([]SqlExpr, len(a))
	for i := range a {
		exprList[i] = a[i]
	}
	return MustJoinExpr(" ", exprList...)
}

func (a Additions) Len() int {
	return len(a)
}

func (a Additions) Less(i, j int) bool {
	return a[i].weight() < a[j].weight()
}

func (a Additions) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type additionWeight int

const (
	// TODO support
	joinStmt additionWeight = iota
	whereStmt
	groupByStmt
	orderByStmt
	limitStmt
	otherStmt
	commentStmt
)

func Join(table *Table, modifies ...string) *join {
	return &join{
		modifies: modifies,
		table:    table,
	}
}

type join struct {
	modifies []string
	table    *Table
	on       *Condition
	using    SqlExpr
}

func (join) weight() additionWeight {
	return joinStmt
}

func (j join) On(c *Condition) *join {
	j.on = c
	return &j
}

func (j join) Using(using SqlExpr) *join {
	j.using = using
	return &j
}

func Where(c *Condition) *where {
	return (*where)(c)
}

type where Condition

func (where) weight() additionWeight {
	return whereStmt
}

func (w *where) Expr() *Expression {
	if w == nil || w.Query == "" {
		return nil
	}
	return Expr("WHERE "+w.Query, w.Args...)
}

func GroupBy(groups ...SqlExpr) *groupBy {
	return &groupBy{
		groups: groups,
	}
}

type groupBy struct {
	groups []SqlExpr
	// WITH ROLLUP
	withRollup bool
	// HAVING
	havingCond *Condition
}

func (groupBy) weight() additionWeight {
	return groupByStmt
}

func (g groupBy) WithRollup() *groupBy {
	g.withRollup = true
	return &g
}

func (g groupBy) Having(cond *Condition) *groupBy {
	g.havingCond = cond
	return &g
}

func (g *groupBy) Expr() *Expression {
	if g == nil {
		return nil
	}
	if len(g.groups) > 0 {
		expr := Expr("GROUP BY")
		for i, order := range g.groups {
			if i == 0 {
				expr = MustJoinExpr(" ", expr, order)
			} else {
				expr = MustJoinExpr(", ", expr, order)
			}
		}

		if g.withRollup {
			expr = MustJoinExpr(" ", expr, Expr("WITH ROLLUP"))
		}

		if g.havingCond != nil {
			expr = MustJoinExpr(" HAVING ", expr, g.havingCond)
		}

		return expr
	}
	return nil
}

func OrderBy(orders ...*order) *orderBy {
	return &orderBy{
		orders: orders,
	}
}

type orderBy struct {
	orders []*order
}

func (orderBy) weight() additionWeight {
	return orderByStmt
}

func (o *orderBy) Expr() *Expression {
	if o == nil {
		return nil
	}
	if len(o.orders) > 0 {
		expr := Expr("ORDER BY")
		for i, order := range o.orders {
			if i == 0 {
				expr = MustJoinExpr(" ", expr, order)
			} else {
				expr = MustJoinExpr(", ", expr, order)
			}
		}
		return expr
	}
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

type order struct {
	expr SqlExpr
	typ  string
}

func (o order) Expr() *Expression {
	if o.expr == nil {
		return nil
	}
	expr := o.expr.Expr()
	if expr == nil {
		return nil
	}
	suffix := ""
	if o.typ != "" {
		suffix = " " + string(o.typ)
	}
	return Expr(
		fmt.Sprintf("(%s)%s", expr.Query, suffix),
		expr.Args...,
	)
}

func Limit(rowCount int32) *limit {
	return &limit{rowCount: rowCount}
}

type limit struct {
	// LIMIT
	rowCount int32
	// OFFSET
	offsetCount int32
}

func (limit) weight() additionWeight {
	return limitStmt
}

func (l limit) Offset(offset int32) *limit {
	l.offsetCount = offset
	return &l
}

func (l *limit) Expr() *Expression {
	if l == nil {
		return nil
	}
	if l.rowCount > 0 {
		if l.offsetCount > 0 {
			return Expr(fmt.Sprintf("LIMIT %d OFFSET %d", l.rowCount, l.offsetCount))
		}
		return Expr(fmt.Sprintf(fmt.Sprintf("LIMIT %d", l.rowCount)))
	}
	return nil
}

type otherAddition Expression

func (otherAddition) weight() additionWeight {
	return otherStmt
}

func (o otherAddition) Expr() *Expression {
	return (*Expression)(&o)
}

func Comment(c string) comment {
	return comment(c)
}

type comment string

func (comment) weight() additionWeight {
	return commentStmt
}

func (c comment) Expr() *Expression {
	return Expr(fmt.Sprintf("/* %s */", string(c)))
}
