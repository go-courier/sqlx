package builder

import (
	"container/list"
	"fmt"
)

type Condition Expression

func NewCondRules() *CondRules {
	return &CondRules{}
}

type CondRules struct {
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

func (rules *CondRules) ToCond() *Condition {
	if rules.l == nil {
		return nil
	}

	list := make([]*Condition, 0)
	i := 0
	for e := rules.l.Front(); e != nil; e = e.Next() {
		r := e.Value.(*CondRule)
		if r.When {
			list = append(list, r.Conditions...)
		}
		i++
	}
	if len(list) == 0 {
		return nil
	}
	return And(list...)
}

func (c *Condition) Expr() *Expression {
	return (*Expression)(c)
}

func (c Condition) And(cond *Condition) *Condition {
	return (*Condition)(Expr(
		fmt.Sprintf("(%s) AND (%s)", c.Query, cond.Query),
		append(c.Args, cond.Args...)...,
	))
}

func (c Condition) Or(cond *Condition) *Condition {
	return (*Condition)(Expr(
		fmt.Sprintf("(%s) OR (%s)", c.Query, cond.Query),
		append(c.Args, cond.Args...)...,
	))
}

func (c Condition) Xor(cond *Condition) *Condition {
	return (*Condition)(Expr(
		fmt.Sprintf("(%s) XOR (%s)", c.Query, cond.Query),
		append(c.Args, cond.Args...)...,
	))
}

func And(condList ...*Condition) (cond *Condition) {
	for _, c := range condList {
		if c == nil {
			continue
		}
		if cond == nil {
			cond = c
			continue
		}
		cond = cond.And(c)
	}
	return
}

func Or(condList ...*Condition) (cond *Condition) {
	for _, c := range condList {
		if c == nil {
			continue
		}
		if cond == nil {
			cond = c
			continue
		}
		cond = cond.Or(c)
	}
	return
}

func Xor(condList ...*Condition) (cond *Condition) {
	for _, c := range condList {
		if c == nil {
			continue
		}
		if cond == nil {
			cond = c
			continue
		}
		cond = cond.Xor(c)
	}
	return
}
