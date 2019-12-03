package builder_test

import (
	"testing"

	. "github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
)

func TestWithStmt(t *testing.T) {
	gr := &GroupRelation{}
	g := &Group{}

	t.Run("simple with", testingx.It(func(t *testingx.T) {
		t.Expect(
			With((&GroupWithParent{}).T()).
				As(func(tmpTableGroupWithParent *Table) SqlExpr {
					s := Select(MultiMayAutoAlias(
						g.T().Col("f_group_id"),
						gr.T().Col("f_group_id"),
					)).
						From(gr.T(),
							RightJoin(g.T()).On(g.T().Col("f_group_id").Eq(gr.T().Col("f_group_id"))),
						)
					return s
				}).
				Do(func(tmpTableGroupWithParent *Table) SqlExpr {
					return Select(nil).From(tmpTableGroupWithParent)
				}),
		).To(buidertestingutils.BeExpr(`
WITH t_group_with_parent(f_group_id,f_parent_group_id) AS (
SELECT (t_group.f_group_id) AS f_group_id, (t_group_relation.f_group_id) AS f_group_id FROM t_group_relation
RIGHT JOIN t_group ON t_group.f_group_id = t_group_relation.f_group_id
)
SELECT * FROM t_group_with_parent
`))
	}))

	t.Run("WithRecursive", testingx.It(func(t *testingx.T) {
		t.Expect(
			WithRecursive((&GroupWithParentAndChildren{}).T()).
				As(func(tmpTableGroupWithParentAndChildren *Table) SqlExpr {
					return With((&GroupWithParent{}).T()).
						As(func(tmpTableGroupWithParent *Table) SqlExpr {
							s := Select(MultiMayAutoAlias(
								g.T().Col("f_group_id"),
								gr.T().Col("f_parent_group_id"),
							)).
								From(gr.T(), RightJoin(g.T()).On(g.T().Col("f_group_id").Eq(gr.T().Col("f_group_id"))))
							return s
						}).
						Do(func(tmpTableGroupWithParent *Table) SqlExpr {
							return UnionAll(
								Select(MultiMayAutoAlias(
									tmpTableGroupWithParent.Col("f_group_id"),
									tmpTableGroupWithParent.Col("f_parent_group_id"),
									Alias(Expr("0"), "f_depth"),
								)).From(
									tmpTableGroupWithParent,
									Where(tmpTableGroupWithParent.Col("f_group_id").Eq(1201375536060956676)),
								),
								Select(MultiMayAutoAlias(
									tmpTableGroupWithParent.Col("f_group_id"),
									tmpTableGroupWithParent.Col("f_parent_group_id"),
									Alias(tmpTableGroupWithParentAndChildren.Col("f_depth").Expr("# + 1"), "f_depth"),
								)).From(
									tmpTableGroupWithParent,
									CrossJoin(tmpTableGroupWithParentAndChildren),
									Where(
										And(
											tmpTableGroupWithParent.Col("f_group_id").Neq(tmpTableGroupWithParentAndChildren.Col("f_group_id")),
											tmpTableGroupWithParent.Col("f_parent_group_id").Eq(tmpTableGroupWithParentAndChildren.Col("f_group_id")),
										)),
								),
							)
						})
				}).
				Do(func(t *Table) SqlExpr {
					return Select(nil).From(t)
				}),
		).To(buidertestingutils.BeExpr(`
WITH RECURSIVE t_group_with_parent_and_children(f_group_id,f_parent_group_id,f_depth) AS (
WITH t_group_with_parent(f_group_id,f_parent_group_id) AS (
SELECT (t_group.f_group_id) AS f_group_id, (t_group_relation.f_parent_group_id) AS f_parent_group_id FROM t_group_relation
RIGHT JOIN t_group ON t_group.f_group_id = t_group_relation.f_group_id
)
SELECT f_group_id, f_parent_group_id, (0) AS f_depth FROM t_group_with_parent
WHERE f_group_id = ?
UNION ALL
SELECT (t_group_with_parent.f_group_id) AS f_group_id, (t_group_with_parent.f_parent_group_id) AS f_parent_group_id, (t_group_with_parent_and_children.f_depth + 1) AS f_depth FROM t_group_with_parent
CROSS JOIN t_group_with_parent_and_children
WHERE (t_group_with_parent.f_group_id <> t_group_with_parent_and_children.f_group_id) AND (t_group_with_parent.f_parent_group_id = t_group_with_parent_and_children.f_group_id)
)
SELECT * FROM t_group_with_parent_and_children
`, 1201375536060956676))
	}))
}

var tableGroup = TableFromModel(&Group{})

type Group struct {
	GroupID int `db:"f_group_id"`
}

func (g *Group) TableName() string {
	return "t_group"
}

func (g *Group) T() *Table {
	return tableGroup
}

var tableGroupRelation = TableFromModel(&GroupRelation{})

type GroupRelation struct {
	GroupID       int `db:"f_group_id"`
	ParentGroupID int `db:"f_parent_group_id"`
}

func (g *GroupRelation) TableName() string {
	return "t_group_relation"
}

func (g *GroupRelation) T() *Table {
	return tableGroupRelation
}

var tableGroupWithParent = TableFromModel(&GroupWithParent{})

type GroupWithParent struct {
	GroupID       int `db:"f_group_id"`
	ParentGroupID int `db:"f_parent_group_id"`
}

func (g *GroupWithParent) TableName() string {
	return "t_group_with_parent"
}

func (g *GroupWithParent) T() *Table {
	return tableGroupWithParent
}

var tableGroupWithParentAndChildren = TableFromModel(&GroupWithParentAndChildren{})

type GroupWithParentAndChildren struct {
	GroupWithParent
	Depth int `db:"f_depth"`
}

func (g *GroupWithParentAndChildren) TableName() string {
	return "t_group_with_parent_and_children"
}

func (g *GroupWithParentAndChildren) T() *Table {
	return tableGroupWithParentAndChildren
}
