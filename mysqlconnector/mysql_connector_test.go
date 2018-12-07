package mysqlconnector

import (
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/go-courier/sqlx/v2/builder"
	"github.com/stretchr/testify/require"
)

func TestMysqlConnector(t *testing.T) {
	c := &MysqlConnector{}

	table := builder.T("t",
		builder.Col("F_id").Type(uint64(0), ",autoincrement"),
		builder.Col("F_name").Type("", ",size=128,default=''"),
		builder.Col("F_geo").Type(&Point{}, ""),
		builder.Col("F_created_at").Type(int64(0), ",default='0'"),
		builder.Col("F_updated_at").Type(int64(0), ",default='0'"),
		builder.PrimaryKey(builder.Cols("F_id")),
		builder.UniqueIndex("I_name", builder.Cols("F_name")).Using("BTREE"),
		builder.Index("I_created_at", builder.Cols("F_created_at")).Using("BTREE"),
		builder.Index("I_geo", builder.Cols("F_geo")).Using("SPATIAL"),
	)

	cases := map[string]struct {
		expr   builder.SqlExpr
		expect builder.SqlExpr
	}{
		"CreateDatabase": {
			c.CreateDatabase("db"),
			builder.Expr( /* language=MySQL */ `CREATE DATABASE db;`),
		},
		"DropDatabase": {
			c.DropDatabase("db"),
			builder.Expr( /* language=MySQL */ `DROP DATABASE db;`),
		},
		"AddIndex": {
			c.AddIndex(table.Key("I_name")),
			builder.Expr( /* language=MySQL */ "CREATE UNIQUE INDEX i_name ON t (f_name) USING BTREE;"),
		},
		"AddPrimaryKey": {
			c.AddIndex(table.Key("PRIMARY")),
			builder.Expr( /* language=MySQL */ "ALTER TABLE t ADD PRIMARY KEY (f_id);"),
		},
		"AddSpatialIndex": {
			c.AddIndex(table.Key("I_geo")),
			builder.Expr( /* language=MySQL */ "CREATE SPATIAL INDEX i_geo ON t (f_geo);"),
		},
		"DropIndex": {
			c.DropIndex(table.Key("I_name")),
			builder.Expr( /* language=MySQL */ "DROP INDEX i_name ON t;"),
		},
		"DropPrimaryKey": {
			c.DropIndex(table.Key("PRIMARY")),
			builder.Expr( /* language=MySQL */ "ALTER TABLE t DROP PRIMARY KEY;"),
		},
		"CreateTableIsNotExists": {
			c.CreateTableIsNotExists(table)[0],
			builder.Expr( /* language=MySQL */ `CREATE TABLE IF NOT EXISTS t (
	f_id bigint unsigned NOT NULL AUTO_INCREMENT,
	f_name varchar(128) NOT NULL DEFAULT '',
	f_geo POINT NOT NULL,
	f_created_at bigint NOT NULL DEFAULT '0',
	f_updated_at bigint NOT NULL DEFAULT '0',
	PRIMARY KEY (f_id)
) ENGINE=InnoDB CHARSET=utf8mb4;`),
		},
		"DropTable": {
			c.DropTable(table),
			builder.Expr( /* language=MySQL */ "DROP TABLE IF EXISTS t;"),
		},
		"TruncateTable": {
			c.TruncateTable(table),
			builder.Expr( /* language=MySQL */ "TRUNCATE TABLE t;"),
		},
		"AddColumn": {
			c.AddColumn(table.Col("F_name")),
			builder.Expr( /* language=MySQL */ "ALTER TABLE t ADD COLUMN f_name varchar(128) NOT NULL DEFAULT '';"),
		},
		"ModifyColumn": {
			c.ModifyColumn(table.Col("F_name")),
			builder.Expr( /* language=MySQL */ "ALTER TABLE t MODIFY COLUMN f_name varchar(128) NOT NULL DEFAULT '';"),
		},
		"DropColumn": {
			c.DropColumn(table.Col("F_name")),
			builder.Expr( /* language=MySQL */ "ALTER TABLE t DROP COLUMN f_name;"),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			queryArgsEqual(t, c.expect, c.expr)
		})
	}
}

func queryArgsEqual(t *testing.T, expect builder.SqlExpr, actual builder.SqlExpr) {
	e := builder.ExprFrom(expect)
	a := builder.ExprFrom(actual)

	if e == nil || a == nil {
		require.Equal(t, e, a)
	} else {
		e = e.Flatten()
		a = a.Flatten()

		require.Equal(t, e.Query(), a.Query())
		require.Equal(t, e.Args(), a.Args())
	}
}

type Point struct {
	X float64
	Y float64
}

func (Point) DataType(engine string) string {
	return "POINT"
}

func (Point) ValueEx() string {
	return `ST_GeomFromText(?)`
}

func (p Point) Value() (driver.Value, error) {
	return fmt.Sprintf("POINT(%v %v)", p.X, p.Y), nil
}
