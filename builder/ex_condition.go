package builder

import (
	"container/list"
)

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
	if c.IsNil() {
		if cond.IsNil() {
			return nil
		}
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
		if cond == nil || cond.IsNil() {
			continue
		}
		c = c.And(cond)
	}
	return c
}

func Or(condList ...SqlCondition) *Condition {
	c := (*Condition)(nil)
	for _, cond := range condList {
		if cond == nil || cond.IsNil() {
			continue
		}
		c = c.Or(cond)
	}
	return c
}

func Xor(condList ...SqlCondition) *Condition {
	c := (*Condition)(nil)
	for _, cond := range condList {
		if cond == nil || cond.IsNil() {
			continue
		}
		c = c.Xor(cond)
	}
	return c
}

func NewCondRules() *CondRules {
	return &CondRules{}
}

type CondRules struct {
	SqlConditionMarker
	l *list.List
}

type CondRule struct {
	When       bool
	Conditions []*Condition
}

func (rules *CondRules) When(rule bool, conditions ...*Condition) *CondRules {
	if rules.l == nil {
		rules.l = list.New()
	}
	rules.l.PushBack(&CondRule{
		When:       rule,
		Conditions: conditions,
	})
	return rules
}

func (rules *CondRules) IsNil() bool {
	if rules == nil || rules.l == nil || rules.l.Len() == 0 {
		return true
	}

	for e := rules.l.Front(); e != nil; e = e.Next() {
		r := e.Value.(*CondRule)
		if r.When {
			return false
		}
	}

	return true
}

func (rules *CondRules) Expr() *Ex {
	if rules.IsNil() {
		return nil
	}

	c := AsCond(Expr(""))

	for e := rules.l.Front(); e != nil; e = e.Next() {
		r := e.Value.(*CondRule)
		if r.When {
			for i := range r.Conditions {
				c = c.And(r.Conditions[i])
			}
		}
	}

	if c.IsNil() {
		return nil
	}

	return c.Expr()
}
