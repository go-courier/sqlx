package sqlx

import (
	"github.com/go-courier/sqlx/builder"
)

type SqlMetaEnum struct {
	TName string `db:"F_table_name" sql:"varchar(64) NOT NULL"`
	CName string `db:"F_column_name" sql:"varchar(64) NOT NULL"`
	Value int    `db:"F_value" sql:"int NOT NULL"`
	Type  string `db:"F_type" sql:"varchar(255) NOT NULL"`
	Key   string `db:"F_key"  sql:"varchar(255) NOT NULL"`
	Label string `db:"F_label" sql:"varchar(255) NOT NULL"`
}

func (*SqlMetaEnum) TableName() string {
	return "t_sql_meta_enum"
}

func (*SqlMetaEnum) UniqueIndexes() builder.Indexes {
	return builder.Indexes{"I_enum": builder.FieldNames{"TName", "CName", "Value"}}
}

func SyncEnum(database *Database, db *DB) error {
	task := NewTasks(db)

	metaEnumTable := database.T(&SqlMetaEnum{})

	task = task.With(func(db *DB) error {
		_, err := db.ExecExpr(builder.CreateTableIsNotExists(metaEnumTable))
		return err
	})

	task = task.With(func(db *DB) error {
		_, err := db.ExecExpr(
			builder.Delete().From(
				metaEnumTable,
				builder.Where(
					metaEnumTable.F("TName").In(database.TableNames()),
				),
			),
		)
		return err
	})

	stmtForInsert := builder.Insert().Into(metaEnumTable)
	vals := make([]interface{}, 0)

	for _, table := range database.Tables {
		table.Columns.Range(func(col *builder.Column, idx int) {
			if col.IsEnum() {
				for _, enum := range col.ColumnType.Enum.ConstValues() {

					sqlMetaEnum := &SqlMetaEnum{
						TName: table.Name,
						CName: col.Name,
						Type:  enum.TypeName(),
						Value: enum.Int(),
						Key:   enum.String(),
						Label: enum.Label(),
					}

					fieldValues := builder.FieldValuesFromStructByNonZero(sqlMetaEnum, "Value")
					_, values := metaEnumTable.ColumnsAndValuesByFieldValues(fieldValues)
					vals = append(vals, values...)
				}
			}
		})
	}

	if len(vals) > 0 {
		stmtForInsert = stmtForInsert.Values(metaEnumTable.Columns, vals...)
		task = task.With(func(db *DB) error {
			_, err := db.ExecExpr(stmtForInsert)
			return err
		})
	}

	return task.Do()
}
