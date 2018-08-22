package generator

import (
	"sort"

	"github.com/go-courier/codegen"
	"github.com/go-courier/packagesx"
	"github.com/go-courier/sqlx/builder"
)

func (m *Model) IndexFieldNames() []string {
	indexedFields := make([]string, 0)

	m.Table.Keys.Range(func(key *builder.Key, idx int) {
		fieldNames := key.Columns.FieldNames()
		indexedFields = append(indexedFields, fieldNames...)
	})

	indexedFields = stringUniq(indexedFields)

	indexedFields = stringFilter(indexedFields, func(item string, i int) bool {
		if m.HasSoftDelete {
			return item != m.FieldKeySoftDelete
		}
		return true
	})

	sort.Strings(indexedFields)

	return indexedFields
}

func (m *Model) WriteTableInterfaces(file *codegen.File) {
	if m.WithTableInterfaces {
		file.WriteBlock(
			codegen.DeclVar(
				codegen.Var(
					codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/builder", "Table"))),
					m.VarTable(),
				),
			),

			codegen.Func().
				Named("init").
				Do(
					codegen.Expr("? = ?.Register(&?{})",
						codegen.Id(m.VarTable()),
						codegen.Id(m.Database),
						codegen.Id(m.StructName),
					),
				),
		)
		file.WriteBlock(
			codegen.Func().
				Named("TableName").
				MethodOf(codegen.Var(m.Type())).
				Return(codegen.Var(codegen.String)).
				Do(
					codegen.Return(file.Val(m.Config.TableName)),
				),
		)

		file.WriteBlock(
			codegen.Func().
				Named("D").
				MethodOf(codegen.Var(m.Type())).
				Return(codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx", "Database"))))).
				Do(
					codegen.Return(codegen.Id(m.Database)),
				),
		)

		file.WriteBlock(
			codegen.Func().
				Named("T").
				MethodOf(codegen.Var(m.Type())).
				Return(codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/builder", "Table"))))).
				Do(
					codegen.Return(codegen.Id(m.VarTable())),
				),
		)
	}

	m.Columns.Range(func(col *builder.Column, idx int) {
		file.WriteBlock(
			codegen.Func().
				Named("FieldKey" + col.FieldName).
				MethodOf(codegen.Var(m.Type())).
				Return(codegen.Var(codegen.String)).
				Do(
					codegen.Return(file.Val(col.FieldName)),
				),
		)

		file.WriteBlock(
			codegen.Func().
				Named("Field" + col.FieldName).
				MethodOf(codegen.Var(m.PtrType(), "m")).
				Return(codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/builder", "Column"))))).
				Do(
					codegen.Return(codegen.Expr("m.T().F(m.FieldKey" + col.FieldName + "())")),
				),
		)
	})

	file.WriteBlock(
		codegen.Func().
			Named("IndexFieldNames").
			MethodOf(codegen.Var(m.PtrType(), "m")).
			Return(codegen.Var(codegen.Slice(codegen.String))).
			Do(
				codegen.Return(file.Val(m.IndexFieldNames())),
			),
	)

	file.WriteBlock(
		codegen.Func().
			Named("ConditionByStruct").
			MethodOf(codegen.Var(m.PtrType(), "m")).
			Return(codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/builder", "Condition"))))).
			Do(
				codegen.Expr(`table := m.T()`),
				codegen.Expr(`fieldValues := `+file.Use("github.com/go-courier/sqlx/builder", "FieldValuesFromStructByNonZero")+`(m)

conditions := make([]*`+file.Use("github.com/go-courier/sqlx/builder", "Condition")+`, 0)

for _, fieldName := range m.IndexFieldNames() {
	if v, exists := fieldValues[fieldName]; exists {
		conditions = append(conditions, table.F(fieldName).Eq(v))
		delete(fieldValues, fieldName)
	}
}

if len(conditions) == 0 {
	panic(`+file.Use("fmt", "Errorf")+`("at least one of field for indexes has value"))
}
	
for fieldName, v := range fieldValues {
	conditions = append(conditions, table.F(fieldName).Eq(v))
}
	
condition := `+file.Use("github.com/go-courier/sqlx/builder", "And")+`(conditions...)
`),

				func() codegen.Snippet {
					if m.HasSoftDelete {
						return codegen.Expr(
							`condition = `+file.Use("github.com/go-courier/sqlx/builder", "And")+`(condition, table.F(?).Eq(?))`,
							file.Val(m.FieldKeySoftDelete),
							file.Use(packagesx.GetPkgImportPathAndExpose(m.ConstSoftDeleteTrue)),
						)
					}
					return codegen.Expr("")
				}(),

				codegen.Return(codegen.Id("condition")),
			),
	)
}

func (m *Model) WriteTableKeyInterfaces(file *codegen.File) {
	if len(m.Keys.Primary) > 0 {
		file.WriteBlock(
			codegen.Func().
				Named("PrimaryKey").
				MethodOf(codegen.Var(m.Type())).
				Return(codegen.Var(codegen.Type(file.Use("github.com/go-courier/sqlx/builder", "FieldNames")))).
				Do(
					codegen.Return(file.Val(m.Keys.Primary)),
				),
		)
	}

	if len(m.Keys.Indexes) > 0 {
		file.WriteBlock(
			codegen.Func().
				Named("Indexes").
				MethodOf(codegen.Var(m.Type())).
				Return(codegen.Var(codegen.Type(file.Use("github.com/go-courier/sqlx/builder", "Indexes")))).
				Do(
					codegen.Return(file.Val(m.Keys.Indexes)),
				),
		)
	}

	if len(m.Keys.UniqueIndexes) > 0 {
		file.WriteBlock(
			codegen.Func().
				Named("UniqueIndexes").
				MethodOf(codegen.Var(m.Type())).
				Return(codegen.Var(codegen.Type(file.Use("github.com/go-courier/sqlx/builder", "Indexes")))).
				Do(
					codegen.Return(file.Val(m.Keys.UniqueIndexes)),
				),
		)
	}

	if m.WithComments {
		file.WriteBlock(
			codegen.Func().
				Named("Comments").
				MethodOf(codegen.Var(m.Type())).
				Return(codegen.Var(codegen.Map(codegen.String, codegen.String))).
				Do(
					codegen.Return(file.Val(m.GetComments())),
				),
		)
	}
}

func (m *Model) GetComments() map[string]string {
	comments := map[string]string{}
	m.Columns.Range(func(col *builder.Column, idx int) {
		comments[col.FieldName] = col.Comment
	})
	return comments
}
