package enummeta

import (
	"github.com/go-courier/enumeration"
	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/builder"
	typex "github.com/go-courier/x/types"
)

func SyncEnum(db sqlx.DBExecutor) error {
	metaEnumTable := builder.T((&SqlMetaEnum{}).TableName())

	dialect := db.Dialect()

	builder.ScanDefToTable(metaEnumTable, &SqlMetaEnum{})

	task := sqlx.NewTasks(db.WithSchema(""))

	task = task.With(func(db sqlx.DBExecutor) error {
		_, err := db.ExecExpr(dialect.DropTable(metaEnumTable))
		return err
	})

	exprs := dialect.CreateTableIsNotExists(metaEnumTable)

	for i := range exprs {
		expr := exprs[i]
		task = task.With(func(db sqlx.DBExecutor) error {
			_, err := db.ExecExpr(expr)
			return err
		})
	}

	{
		// insert values
		stmtForInsert := builder.Insert().Into(metaEnumTable)
		vals := make([]interface{}, 0)

		columns := &builder.Columns{}

		db.D().Tables.Range(func(table *builder.Table, idx int) {
			table.Columns.Range(func(col *builder.Column, idx int) {
				if rv, ok := typex.TryNew(col.ColumnType.Type); ok {
					if enumValue, ok := rv.Interface().(enumeration.Enum); ok {
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
				}
			})
		})

		if len(vals) > 0 {
			stmtForInsert = stmtForInsert.Values(columns, vals...)

			task = task.With(func(db sqlx.DBExecutor) error {
				_, err := db.ExecExpr(stmtForInsert)
				return err
			})
		}
	}

	return task.Do()
}
