package enummeta

import (
	"reflect"

	"github.com/go-courier/enumeration"
	"github.com/go-courier/sqlx/v2/builder"

	"github.com/go-courier/sqlx/v2"
)

func SyncEnum(db *sqlx.DB, database *sqlx.Database) error {
	task := sqlx.NewTasks(db)

	metaEnumTable := database.T(&SqlMetaEnum{})

	task = task.With(func(db *sqlx.DB) error {
		_, err := db.ExecExpr(
			builder.Delete().From(
				metaEnumTable,
				builder.Where(metaEnumTable.F("TName").In(database.Tables.TableNames())),
			),
		)
		return err
	})

	stmtForInsert := builder.Insert().Into(metaEnumTable)
	vals := make([]interface{}, 0)

	columns := &builder.Columns{}

	for _, table := range database.Tables {
		table.Columns.Range(func(col *builder.Column, idx int) {
			v := reflect.New(col.ColumnType.Type).Interface()
			if enumValue, ok := v.(enumeration.Enum); ok {
				for _, enum := range enumValue.ConstValues() {
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
