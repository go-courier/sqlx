package builder

type SqlCondition interface {
	SqlExpr
	SqlConditionMarker
}

type SqlConditionMarker interface {
	asCondition()
}

func AsCond(ex SqlExpr) *Condition {
	if ex.IsNil() {
		return &Condition{SqlExpr: Expr("")}
	}
	return &Condition{SqlExpr: ex}
}

type Condition struct {
	SqlExpr
	SqlConditionMarker
}

func (c *Condition) IsNil() bool {
	return c == nil || c.SqlExpr.IsNil()
}

func (c *Condition) And(cond SqlCondition) *Condition {
	if cond == nil || cond.IsNil() {
		return c
	}

	if c.IsNil() {
		return AsCond(cond.Expr())
	}

	e := Expr("")
	e.WriteGroup(func(e *Ex) {
		e.WriteExpr(c)
	})
	e.WriteString(" AND ")
	e.WriteGroup(func(e *Ex) {
		e.WriteExpr(cond)
	})
	return AsCond(e)
}

func (c *Condition) Or(cond SqlCondition) *Condition {
	if cond == nil || cond.IsNil() {
		return c
	}

	if c.IsNil() {
		return AsCond(cond.Expr())
	}

	e := Expr("")
	e.WriteGroup(func(e *Ex) {
		e.WriteExpr(c)
	})
	e.WriteString(" OR ")
	e.WriteGroup(func(e *Ex) {
		e.WriteExpr(cond)
	})
	return AsCond(e)
}

func (c *Condition) Xor(cond SqlCondition) *Condition {
	if cond == nil || cond.IsNil() {
		return c
	}

	if c.IsNil() {
		return AsCond(cond.Expr())
	}

	e := Expr("")
	e.WriteGroup(func(e *Ex) {
		e.WriteExpr(c)
	})
	e.WriteString(" XOR ")
	e.WriteGroup(func(e *Ex) {
		e.WriteExpr(cond)
	})
	return AsCond(e)
}

func And(condList ...SqlCondition) *Condition {
	c := (*Condition)(nil)
	for _, cond := range condList {
		c = c.And(cond)
	}
	return c
}

func Or(condList ...SqlCondition) *Condition {
	c := (*Condition)(nil)
	for _, cond := range condList {
		c = c.Or(cond)
	}
	return c
}

func Xor(condList ...SqlCondition) *Condition {
	c := (*Condition)(nil)
	for _, cond := range condList {
		c = c.Xor(cond)
	}
	return c
}
