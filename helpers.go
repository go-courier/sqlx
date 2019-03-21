package sqlx

import "github.com/go-courier/sqlx/v2/builder"

func InsertToDB(db DBExecutor, model builder.Model, zeroFields []string, additions ...builder.Addition) builder.SqlExpr {
	table := db.T(model)
	cols, vals := table.ColumnsAndValuesByFieldValues(FieldValuesFromModel(table, model, zeroFields...))
	return builder.Insert().Into(table, additions...).Values(cols, vals...)
}

func AsAssignments(db DBExecutor, model builder.Model, zeroFields ...string) builder.Assignments {
	table := db.T(model)
	return table.AssignmentsByFieldValues(FieldValuesFromModel(table, model, zeroFields...))
}

func FieldValuesFromModel(table *builder.Table, model builder.Model, zeroFields ...string) builder.FieldValues {
	fieldValues := builder.FieldValuesFromStructByNonZero(model, zeroFields...)
	if autoIncrementCol := table.AutoIncrement(); autoIncrementCol != nil {
		delete(fieldValues, autoIncrementCol.FieldName)
	}
	return fieldValues
}
