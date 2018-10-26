package mysqlconnector

import (
	"database/sql/driver"
	"fmt"
	"github.com/go-courier/sqlx/builder"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMysqlDialect(t *testing.T) {
	d := &MysqlConnector{}

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
		"CreateDatabaseIfNotExists": {
			d.CreateDatabaseIfNotExists("db"),
			builder.Expr( /* language=MySQL */ `CREATE DATABASE IF NOT EXISTS db;`),
		},
		"DropDatabase": {
			d.DropDatabase("db"),
			builder.Expr( /* language=MySQL */ `DROP DATABASE db;`),
		},
		"AddIndex": {
			d.AddIndex(table.Key("I_name")),
			builder.Expr( /* language=MySQL */ "CREATE UNIQUE INDEX I_name ON t (F_name) USING BTREE;"),
		},
		"AddPrimaryKey": {
			d.AddIndex(table.Key("PRIMARY")),
			builder.Expr( /* language=MySQL */ "ALTER TABLE t ADD PRIMARY KEY (F_id);"),
		},
		"AddSpatialIndex": {
			d.AddIndex(table.Key("I_geo")),
			builder.Expr( /* language=MySQL */ "CREATE SPATIAL INDEX I_geo ON t (F_geo);"),
		},
		"DropIndex": {
			d.DropIndex(table.Key("I_name")),
			builder.Expr( /* language=MySQL */ "DROP INDEX I_name ON t;"),
		},
		"DropPrimaryKey": {
			d.DropIndex(table.Key("PRIMARY")),
			builder.Expr( /* language=MySQL */ "ALTER TABLE t DROP PRIMARY KEY;"),
		},
		"CreateTableIsNotExists": {
			d.CreateTableIsNotExists(table)[0],
			builder.Expr( /* language=MySQL */ `CREATE TABLE IF NOT EXISTS t (
	F_id bigint unsigned NOT NULL AUTO_INCREMENT,
	F_name varchar(128) NOT NULL DEFAULT '',
	F_geo POINT NOT NULL,
	F_created_at bigint NOT NULL DEFAULT '0',
	F_updated_at bigint NOT NULL DEFAULT '0',
	PRIMARY KEY (F_id)
) ENGINE=InnoDB CHARSET=utf8mb4;`),
		},
		"DropTable": {
			d.DropTable(table),
			builder.Expr( /* language=MySQL */ "DROP TABLE t;"),
		},
		"TruncateTable": {
			d.TruncateTable(table),
			builder.Expr( /* language=MySQL */ "TRUNCATE TABLE t;"),
		},
		"AddColumn": {
			d.AddColumn(table.Col("F_name")),
			builder.Expr( /* language=MySQL */ "ALTER TABLE t ADD COLUMN F_name varchar(128) NOT NULL DEFAULT '';"),
		},
		"ModifyColumn": {
			d.ModifyColumn(table.Col("F_name")),
			builder.Expr( /* language=MySQL */ "ALTER TABLE t MODIFY COLUMN F_name varchar(128) NOT NULL DEFAULT '';"),
		},
		"DropColumn": {
			d.DropColumn(table.Col("F_name")),
			builder.Expr( /* language=MySQL */ "ALTER TABLE t DROP COLUMN F_name;"),
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
