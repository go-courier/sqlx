package mysqlconnector

import (
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/sqlx/v2/builder/buidertestingutils"
	"github.com/go-courier/testingx"
)

func TestMysqlConnector(t *testing.T) {
	c := &MysqlConnector{}

	table := builder.T("t",
		builder.Col("F_id").Type(uint64(0), ",autoincrement"),
		builder.Col("f_old_name").Type("", ",deprecated=f_name"),
		builder.Col("F_name").Type("", ",size=128,default=''"),
		builder.Col("F_geo").Type(&Point{}, ""),
		builder.Col("F_created_at").Type(int64(0), ",default='0'"),
		builder.Col("F_updated_at").Type(int64(0), ",default='0'"),
		builder.PrimaryKey(builder.Cols("F_id")),
		builder.UniqueIndex("I_name", builder.Cols("F_name")).Using("BTREE"),
		builder.Index("I_created_at", builder.Cols("F_created_at")).Using("BTREE"),
		builder.Index("I_geo", builder.Cols("F_geo")).Using("SPATIAL"),
	)

	t.Run("CreateDatabase", testingx.It(func(t *testingx.T) {
		t.Expect(c.CreateDatabase("db")).
			To(buidertestingutils.BeExpr( /* language=MySQL */ `CREATE DATABASE db;`))
	}))

	t.Run("DropDatabase", testingx.It(func(t *testingx.T) {
		t.Expect(c.DropDatabase("db")).
			To(buidertestingutils.BeExpr( /* language=MySQL */ `DROP DATABASE db;`))
	}))

	t.Run("AddIndex", testingx.It(func(t *testingx.T) {
		t.Expect(c.AddIndex(table.Key("I_name"))).
			To(buidertestingutils.BeExpr( /* language=MySQL */ `CREATE UNIQUE INDEX i_name ON t (f_name) USING BTREE;`))
	}))

	t.Run("AddPrimaryKey", testingx.It(func(t *testingx.T) {
		t.Expect(c.AddIndex(table.Key("PRIMARY"))).
			To(buidertestingutils.BeExpr( /* language=MySQL */ "ALTER TABLE t ADD PRIMARY KEY (f_id);"))
	}))

	t.Run("AddSpatialIndex", testingx.It(func(t *testingx.T) {
		t.Expect(
			c.AddIndex(table.Key("I_geo")),
		).To(buidertestingutils.BeExpr( /* language=MySQL */ "CREATE SPATIAL INDEX i_geo ON t (f_geo);"))
	}))

	t.Run("DropIndex", testingx.It(func(t *testingx.T) {
		t.Expect(
			c.DropIndex(table.Key("I_name")),
		).To(buidertestingutils.BeExpr( /* language=MySQL */ "DROP INDEX i_name ON t;"))
	}))
	t.Run("DropPrimaryKey", testingx.It(func(t *testingx.T) {
		t.Expect(
			c.DropIndex(table.Key("PRIMARY")),
		).To(buidertestingutils.BeExpr( /* language=MySQL */ "ALTER TABLE t DROP PRIMARY KEY;"))
	}))

	t.Run("CreateTableIsNotExists", testingx.It(func(t *testingx.T) {
		t.Expect(
			c.CreateTableIsNotExists(table)[0],
		).To(buidertestingutils.BeExpr( /* language=MySQL */
			`CREATE TABLE IF NOT EXISTS t (
	f_id bigint unsigned NOT NULL AUTO_INCREMENT,
	f_name varchar(128) NOT NULL DEFAULT '',
	f_geo POINT NOT NULL,
	f_created_at bigint NOT NULL DEFAULT '0',
	f_updated_at bigint NOT NULL DEFAULT '0',
	PRIMARY KEY (f_id)
) ENGINE=InnoDB CHARSET=utf8mb4;`))
	}))

	t.Run("DropTable", testingx.It(func(t *testingx.T) {
		t.Expect(
			c.DropTable(table)).
			To(buidertestingutils.BeExpr( /* language=MySQL */ "DROP TABLE IF EXISTS t;"))
	}))
	t.Run("TruncateTable", testingx.It(func(t *testingx.T) {
		t.Expect(
			c.TruncateTable(table),
		).
			To(buidertestingutils.BeExpr( /* language=MySQL */ "TRUNCATE TABLE t;"))
	}))
	t.Run("AddColumn", testingx.It(func(t *testingx.T) {
		t.Expect(
			c.AddColumn(table.Col("F_name"))).
			To(buidertestingutils.BeExpr( /* language=MySQL */ "ALTER TABLE t ADD COLUMN f_name varchar(128) NOT NULL DEFAULT '';"))
	}))

	t.Run("ModifyColumn", testingx.It(func(t *testingx.T) {
		t.Expect(
			c.ModifyColumn(table.Col("F_name")),
		).To(buidertestingutils.BeExpr( /* language=MySQL */ "ALTER TABLE t MODIFY COLUMN f_name varchar(128) NOT NULL DEFAULT '';"))
	}))

	t.Run("DropColumn", testingx.It(func(t *testingx.T) {
		t.Expect(
			c.DropColumn(table.Col("F_name")),
		).To(buidertestingutils.BeExpr( /* language=MySQL */ "ALTER TABLE t DROP COLUMN f_name;"))
	}))
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
