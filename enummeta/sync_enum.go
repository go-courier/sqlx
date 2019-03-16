package enummeta

import (
	"github.com/go-courier/enumeration"
	"reflect"

	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/builder"
)

func SyncEnum(db *sqlx.DB) error {
	metaEnumTable := builder.T((&SqlMetaEnum{}).TableName())
	builder.ScanDefToTable(reflect.ValueOf(&SqlMetaEnum{}), metaEnumTable)

	task := sqlx.NewTasks(db.WithSchema(""))

	task = task.With(func(db *sqlx.DB) error {
		_, err := db.ExecExpr(db.DropTable(metaEnumTable))
		return err
	})

	exprs := db.CreateTableIsNotExists(metaEnumTable)

	for i := range exprs {
		expr := exprs[i]
		task = task.With(func(db *sqlx.DB) error {
			_, err := db.ExecExpr(expr)
			return err
		})
	}

	{
		// insert values
		stmtForInsert := builder.Insert().Into(metaEnumTable)
		vals := make([]interface{}, 0)

		columns := &builder.Columns{}

		db.Tables.Range(func(table *builder.Table, idx int) {
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
		})

		if len(vals) > 0 {
			stmtForInsert = stmtForInsert.Values(columns, vals...)

			task = task.With(func(db *sqlx.DB) error {
				_, err := db.ExecExpr(stmtForInsert)
				return err
			})
		}
	}

	return task.Do()
}
