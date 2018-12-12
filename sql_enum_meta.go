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
	metaEnumTable := database.T(&SqlMetaEnum{})

	if _, err := db.ExecExpr(builder.CreateTableIsNotExists(metaEnumTable)); err != nil {
		return err
	}

	if _, err := db.ExecExpr(
		builder.Delete().From(
			metaEnumTable,
			builder.Where(
				metaEnumTable.F("TName").In(database.TableNames()),
			),
		),
	); err != nil {
		return err
	}

	stmtForInsert := builder.Insert().Into(metaEnumTable)
	vals := make([]interface{}, 0)

	columns := &builder.Columns{}

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
					cols, values := metaEnumTable.ColumnsAndValuesByFieldValues(fieldValues)
					vals = append(vals, values...)
					columns = cols
				}
			}
		})
	}

	if len(vals) > 0 {
		stmtForInsert = stmtForInsert.Values(columns, vals...)

		if _, err := db.ExecExpr(stmtForInsert); err != nil {
			return err
		}
	}

	return nil
}
