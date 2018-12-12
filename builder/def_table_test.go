package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTable(t *testing.T) {
	tt := require.New(t)

	db := DB("db")
	table := T(db, "t")

	table = table.Define(
		Col(table, "F_id").Field("ID"), // skip without type
		Col(table, "F_id").Field("ID").Type("bigint(64) unsigned NOT NULL AUTO_INCREMENT"),
		Col(table, "F_name").Field("Name").Type("varchar(255) NOT NULL DEFAULT ''"),
		Col(table, "F_username").Field("Username").Type("varchar(255) NOT NULL DEFAULT ''"),
		Col(table, "F_created_at").Field("CreatedAt").Type("bigint(64) NOT NULL DEFAULT '0'"),
		Col(table, "F_updated_at").Field("UpdatedAt").Type("bigint(64) NOT NULL DEFAULT '0'"),
		PrimaryKey(), // skip without Columns
		PrimaryKey().WithCols(Col(table, "F_id")),
		Index("I_name").WithCols(Col(table, "F_name")),
		UniqueIndex("I_username").WithCols(Col(table, "F_name"), Col(table, "F_id")),
	)

	{
		cols, values := table.ColumnsAndValuesByFieldValues(FieldValues{
			"ID":   1,
			"Name": "1",
		})

		tt.Equal(table.Fields("ID", "Name").Len(), cols.Len())
		if values[0] == 1 {
			tt.Equal([]interface{}{1, "1"}, values)
		} else {
			tt.Equal([]interface{}{"1", 1}, values)
		}
	}

	{
		assignments := table.AssignmentsByFieldValues(FieldValues{
			"ID":   1,
			"Name": "1",
		})

		expr := assignments.Expr()

		if expr.Args[0] == 1 {
			tt.Equal(Expr("`F_id` = ?, `F_name` = ?", 1, "1"), assignments.Expr())
		} else {
			tt.Equal(Expr("`F_name` = ?, `F_id` = ?", "1", 1), assignments.Expr())
		}

	}

	cases := []struct {
		name   string
		expr   SqlExpr
		result SqlExpr
	}{
		{
			"create Table",
			CreateTableIsNotExists(table),
			Expr("CREATE TABLE IF NOT EXISTS `db`.`t` (" +
				"`F_id` bigint(64) unsigned NOT NULL AUTO_INCREMENT, " +
				"`F_name` varchar(255) NOT NULL DEFAULT '', " +
				"`F_username` varchar(255) NOT NULL DEFAULT '', " +
				"`F_created_at` bigint(64) NOT NULL DEFAULT '0', " +
				"`F_updated_at` bigint(64) NOT NULL DEFAULT '0', " +
				"PRIMARY KEY (`F_id`), " +
				"INDEX `I_name` (`F_name`), " +
				"UNIQUE INDEX `I_username` (`F_name`,`F_id`)" +
				") ENGINE=InnoDB CHARSET=utf8"),
		}, {
			"create Table",
			CreateTable(table),
			Expr("CREATE TABLE `db`.`t` (" +
				"`F_id` bigint(64) unsigned NOT NULL AUTO_INCREMENT, " +
				"`F_name` varchar(255) NOT NULL DEFAULT '', " +
				"`F_username` varchar(255) NOT NULL DEFAULT '', " +
				"`F_created_at` bigint(64) NOT NULL DEFAULT '0', " +
				"`F_updated_at` bigint(64) NOT NULL DEFAULT '0', " +
				"PRIMARY KEY (`F_id`), " +
				"INDEX `I_name` (`F_name`), " +
				"UNIQUE INDEX `I_username` (`F_name`,`F_id`)" +
				") ENGINE=InnoDB CHARSET=utf8"),
		}, {
			"cond",
			table.Cond("#ID = ? AND #Username = ?"),
			Expr("`F_id` = ? AND `F_username` = ?"),
		}, {
			"cond with unregister col field",
			table.Cond("#ID = ? AND #Usernames = ?"),
			Expr("`F_id` = ? AND #Usernames = ?"),
		}, {
			"drop Table",
			DropTable(table),
			Expr("DROP TABLE `db`.`t`"),
		}, {
			"truncate Table",
			TruncateTable(table),
			Expr("TRUNCATE TABLE `db`.`t`"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, ExprFrom(c.result), ExprFrom(c.expr))
		})
	}
}
