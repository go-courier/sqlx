package builder

import (
	"context"
)

func EmptyCond() SqlCondition {
	return (*Condition)(nil)
}

type SqlCondition interface {
	SqlExpr
	SqlConditionMarker

	And(cond SqlCondition) SqlCondition
	Or(cond SqlCondition) SqlCondition
	Xor(cond SqlCondition) SqlCondition
}

type SqlConditionMarker interface {
	asCondition()
}

func AsCond(ex SqlExpr) *Condition {
	return &Condition{expr: ex}
}

type Condition struct {
	expr SqlExpr
	SqlConditionMarker
}

func (c *Condition) Ex(ctx context.Context) *Ex {
	if IsNilExpr(c.expr) {
		return nil
	}
	return c.expr.Ex(ctx)
}

func (c *Condition) IsNil() bool {
	return c == nil || IsNilExpr(c.expr)
}

func (c *Condition) And(cond SqlCondition) SqlCondition {
	return And(c, cond)
}

func (c *Condition) Or(cond SqlCondition) SqlCondition {
	return Or(c, cond)
}

func (c *Condition) Xor(cond SqlCondition) SqlCondition {
	return Xor(c, cond)
}

func And(conditions ...SqlCondition) SqlCondition {
	return composedCondition("AND", conditions...)
}

func Or(conditions ...SqlCondition) SqlCondition {
	return composedCondition("OR", conditions...)
}

func Xor(conditions ...SqlCondition) SqlCondition {
	return composedCondition("XOR", conditions...)
}

func composedCondition(op string, conditions ...SqlCondition) SqlCondition {
	final := filterNilCondition(conditions...)

	if len(final) == 0 {
		return nil
	}

	if len(final) == 1 {
		return final[0]
	}

	return &ComposedCondition{op: op, conditions: final}
}

func filterNilCondition(conditions ...SqlCondition) []SqlCondition {
	finalConditions := make([]SqlCondition, 0)

	for i := range conditions {
		condition := conditions[i]
		if IsNilExpr(condition) {
			continue
		}
		finalConditions = append(finalConditions, condition)
	}

	return finalConditions
}

type ComposedCondition struct {
	op         string
	conditions []SqlCondition
	SqlConditionMarker
}

func (c *ComposedCondition) And(cond SqlCondition) SqlCondition {
	return And(c, cond)
}

func (c *ComposedCondition) Or(cond SqlCondition) SqlCondition {
	return Or(c, cond)
}

func (c *ComposedCondition) Xor(cond SqlCondition) SqlCondition {
	return Xor(c, cond)
}

func (c *ComposedCondition) IsNil() bool {
	return c == nil || c.op == "" || len(c.conditions) == 0
}

func (c *ComposedCondition) Ex(ctx context.Context) *Ex {
	e := Expr("")

	count := 0

	for i := range c.conditions {
		condition := c.conditions[i]
		if condition == nil || condition.IsNil() {
			continue
		}

		if count > 0 {
			e.WriteByte(' ')
			e.WriteString(c.op)
			e.WriteByte(' ')
		}

		e.WriteGroup(func(e *Ex) {
			e.WriteExpr(condition)
		})

		count++
	}

	return e.Ex(ctx)
}
