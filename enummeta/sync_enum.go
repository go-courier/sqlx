package enummeta

import (
	"github.com/go-courier/sqlx"
	"github.com/go-courier/sqlx/builder"
)

func SyncEnum(database *sqlx.Database, db *sqlx.DB) error {
	task := sqlx.NewTasks(db)

	metaEnumTable := database.T(&SqlMetaEnum{})

	task = task.With(func(db *sqlx.DB) error {
		_, err := db.ExecExpr(builder.CreateTableIsNotExists(metaEnumTable))
		return err
	})

	task = task.With(func(db *sqlx.DB) error {
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

		task = task.With(func(db *sqlx.DB) error {
			_, err := db.ExecExpr(stmtForInsert)
			return err
		})
	}

	return task.Do()
}
