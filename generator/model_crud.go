package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-courier/codegen"
	"github.com/go-courier/packagesx"
	"github.com/go-courier/sqlx/builder"
)

func (m *Model) SnippetEnableIfNeed(file *codegen.File) codegen.Snippet {
	if m.HasSoftDelete {
		return codegen.Expr("m.? = ?",
			codegen.Id(m.FieldKeySoftDelete),
			codegen.Id(file.Use(packagesx.GetPkgImportPathAndExpose(m.ConstSoftDeleteTrue))),
		)
	}
	return nil
}

func (m *Model) SnippetDisableIfNeedForFieldValues(file *codegen.File) codegen.Snippet {
	if m.HasSoftDelete {
		return codegen.Expr(`if _, ok := fieldValues["?"]; !ok {
			fieldValues["?"] = ?
		}`,
			codegen.Id(m.FieldKeySoftDelete),
			codegen.Id(m.FieldKeySoftDelete),
			codegen.Id(file.Use(packagesx.GetPkgImportPathAndExpose(m.ConstSoftDeleteFalse))),
		)
	}
	return nil
}

func (m *Model) SnippetEnableIfNeedForFieldValues(file *codegen.File) codegen.Snippet {
	if m.HasSoftDelete {
		return codegen.Expr(`if _, ok := fieldValues["?"]; !ok {
			fieldValues["?"] = ?
		}`,
			codegen.Id(m.FieldKeySoftDelete),
			codegen.Id(m.FieldKeySoftDelete),
			codegen.Id(file.Use(packagesx.GetPkgImportPathAndExpose(m.ConstSoftDeleteTrue))),
		)
	}
	return nil
}

func (m *Model) SnippetSetCreatedAtIfNeed(file *codegen.File) codegen.Snippet {
	if m.HasCreatedAt {
		return codegen.Expr(`
if m.?.IsZero() {
	m.? = ?(?())
}
`,
			codegen.Id(m.FieldKeyCreatedAt),
			codegen.Id(m.FieldKeyCreatedAt),
			codegen.Id(m.FieldType(file, m.FieldKeyCreatedAt)),
			codegen.Id(file.Use("time", "Now")))
	}

	return nil
}

func (m *Model) SnippetSetLastInsertIdIfNeed(file *codegen.File) codegen.Snippet {
	if m.HasAutoIncrement {
		return codegen.Expr(`
if err == nil {
	lastInsertID, _ := result.LastInsertId()
	m.? = `+m.FieldType(file, m.FieldKeyAutoIncrement)+`(lastInsertID)
}
`,
			codegen.Id(m.FieldKeyAutoIncrement),
		)
	}

	return codegen.Expr(`
_ = result
`)
}

func (m *Model) SnippetSetUpdatedAtIfNeed(file *codegen.File) codegen.Snippet {
	if m.HasCreatedAt {
		return codegen.Expr(`
if m.?.IsZero() {
	m.? = ?(?())
}
`,
			codegen.Id(m.FieldKeyUpdatedAt),
			codegen.Id(m.FieldKeyUpdatedAt),
			codegen.Id(m.FieldType(file, m.FieldKeyUpdatedAt)),
			codegen.Id(file.Use("time", "Now")))
	}

	return codegen.Expr("")

	return nil
}

func (m *Model) SnippetSetUpdatedAtIfNeedForFieldValues(file *codegen.File) codegen.Snippet {
	if m.HasAutoIncrement {
		return codegen.Expr(`
if _, ok := fieldValues[?]; !ok {
	fieldValues[?] = ?(?())
}
`,
			codegen.Val(m.FieldKeyUpdatedAt),
			codegen.Val(m.FieldKeyUpdatedAt),
			codegen.Id(m.FieldType(file, m.FieldKeyUpdatedAt)),
			codegen.Id(file.Use("time", "Now")))
	}
	return nil
}

func (m *Model) WriteCreate(file *codegen.File) {
	file.WriteBlock(
		codegen.Func(codegen.Var(
			codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx", "DB"))), "db")).
			Named("Create").
			MethodOf(codegen.Var(m.PtrType(), "m")).
			Return(codegen.Var(codegen.Error)).
			Do(
				m.SnippetEnableIfNeed(file),
				m.SnippetSetCreatedAtIfNeed(file),

				codegen.Expr(`
d := m.D()

result, err := db.ExecExpr(`+file.Use("github.com/go-courier/sqlx/builder", "Insert")+`().
Into(m.T(), `+file.Use("github.com/go-courier/sqlx/builder", "Comment")+`(?)).
Set(d.Assignments(m)...))

`, file.Val(m.StructName+".Create")),

				m.SnippetSetLastInsertIdIfNeed(file),

				codegen.Return(codegen.Expr("err")),
			),
	)

	if len(m.Keys.UniqueIndexes) > 0 {

		file.WriteBlock(
			codegen.Func(
				codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx", "DB"))), "db"),
				codegen.Var(codegen.Slice(codegen.String), "updateFields"),
			).
				Named("CreateOnDuplicateWithUpdateFields").
				MethodOf(codegen.Var(m.PtrType(), "m")).
				Return(codegen.Var(codegen.Error)).
				Do(
					codegen.Expr(`
if len(updateFields) == 0 {
	panic(`+file.Use("fmt", "Errorf")+`("must have update fields"))
}
`),

					m.SnippetEnableIfNeed(file),
					m.SnippetSetCreatedAtIfNeed(file),

					codegen.Expr(`
fieldValues := `+file.Use("github.com/go-courier/sqlx/builder", "FieldValuesFromStructByNonZero")+`(m, updateFields...)
`),
					func() codegen.Snippet {
						if m.HasAutoIncrement {
							return codegen.Expr(
								`delete(fieldValues, ?)`,
								file.Val(m.FieldKeyAutoIncrement),
							)
						}
						return nil
					}(),

					codegen.Expr(`
table := m.T()

cols, vals := table.ColumnsAndValuesByFieldValues(fieldValues)

fields := make(map[string]bool, len(updateFields))
for _, field := range updateFields {
	fields[field] = true
}
`),
					codegen.Expr(`
for _, fieldNames := range m.UniqueIndexes() {
	for _, field := range fieldNames {
		delete(fields, field)
	}
}

if len(fields) == 0 {
	panic(`+file.Use("fmt", "Errorf")+`("no fields for updates"))
}

for field := range fieldValues {
	if !fields[field] {
		delete(fieldValues, field)
	}
}


_, err := db.ExecExpr(`+file.Use("github.com/go-courier/sqlx/builder", "Insert")+`().
	Into(
		m.T(), 
		`+file.Use("github.com/go-courier/sqlx/builder", "OnDuplicateKeyUpdate")+`(table.AssignmentsByFieldValues(fieldValues)...),
		`+file.Use("github.com/go-courier/sqlx/builder", "Comment")+`(?),
	).
	Values(cols, vals...),
)

`, file.Val(m.StructName+".CreateOnDuplicateWithUpdateFields")),

					codegen.Return(codegen.Expr("err")),
				),
		)
	}

}

func (m *Model) WriteDelete(file *codegen.File) {
	file.WriteBlock(
		codegen.Func(codegen.Var(
			codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx", "DB"))), "db")).
			Named("DeleteByStruct").
			MethodOf(codegen.Var(m.PtrType(), "m")).
			Return(codegen.Var(codegen.Error)).
			Do(
				m.SnippetEnableIfNeed(file),
				m.SnippetSetCreatedAtIfNeed(file),

				codegen.Expr(`
_, err := db.ExecExpr(
`+file.Use("github.com/go-courier/sqlx/builder", "Delete")+`().
From(
	m.T(),
`+file.Use("github.com/go-courier/sqlx/builder", "Where")+`(m.ConditionByStruct()),
`+file.Use("github.com/go-courier/sqlx/builder", "Comment")+`(?),
),
)

`, file.Val(m.StructName+".DeleteByStruct")),

				codegen.Return(codegen.Expr("err")),
			),
	)
}

func (m *Model) WriteByKey(file *codegen.File) {
	m.Table.Keys.Range(func(key *builder.Key, idx int) {
		fieldNames := key.Columns.FieldNames()

		fieldNamesWithoutEnabled := stringFilter(fieldNames, func(item string, i int) bool {
			if m.HasSoftDelete {
				return item != m.FieldKeySoftDelete
			}
			return true
		})

		if m.HasSoftDelete && key.Type == builder.PRIMARY {
			fieldNames = append(fieldNames, m.FieldKeySoftDelete)
		}

		if key.Type == builder.PRIMARY || key.Type == builder.UNIQUE_INDEX {
			{
				methodForFetch := createMethod("FetchBy%s", fieldNamesWithoutEnabled...)

				file.WriteBlock(
					codegen.Func(codegen.Var(
						codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx", "DB"))), "db")).
						Named(methodForFetch).
						MethodOf(codegen.Var(m.PtrType(), "m")).
						Return(codegen.Var(codegen.Error)).
						Do(
							m.SnippetEnableIfNeed(file),

							codegen.Expr(`
table := m.T()

err := db.QueryExprAndScan(
`+file.Use("github.com/go-courier/sqlx/builder", "Select")+`(nil).
From(
	m.T(),
`+file.Use("github.com/go-courier/sqlx/builder", "Where")+`(`+toExactlyConditionFrom(file, fieldNames...)+`),
`+file.Use("github.com/go-courier/sqlx/builder", "Comment")+`(?),
),
m,
)
`,
								file.Val(m.StructName+"."+methodForFetch),
							),

							codegen.Return(codegen.Expr("err")),
						),
				)

				methodForUpdateWithMap := createMethod("UpdateBy%sWithMap", fieldNamesWithoutEnabled...)

				file.WriteBlock(
					codegen.Func(
						codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx", "DB"))), "db"),
						codegen.Var(codegen.Type(file.Use("github.com/go-courier/sqlx/builder", "FieldValues")), "fieldValues"),
					).
						Named(methodForUpdateWithMap).
						MethodOf(codegen.Var(m.PtrType(), "m")).
						Return(codegen.Var(codegen.Error)).
						Do(
							m.SnippetSetUpdatedAtIfNeedForFieldValues(file),
							m.SnippetEnableIfNeed(file),

							codegen.Expr(`
table := m.T()

result, err := db.ExecExpr(
	`+file.Use("github.com/go-courier/sqlx/builder", "Update")+`(m.T()).
		Where(
			`+toExactlyConditionFrom(file, fieldNames...)+`,
			`+file.Use("github.com/go-courier/sqlx/builder", "Comment")+`(?),
		).
		Set(table.AssignmentsByFieldValues(fieldValues)...),
	)

if err != nil {
	return err
}

rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
  return m.`+methodForFetch+`(db)
}

return nil
`,
								file.Val(m.StructName+"."+methodForUpdateWithMap),
							),
						),
				)

				methodForUpdateWithStruct := createMethod("UpdateBy%sWithStruct", fieldNamesWithoutEnabled...)

				file.WriteBlock(
					codegen.Func(
						codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx", "DB"))), "db"),
						codegen.Var(codegen.Ellipsis(codegen.String), "zeroFields"),
					).
						Named(methodForUpdateWithStruct).
						MethodOf(codegen.Var(m.PtrType(), "m")).
						Return(codegen.Var(codegen.Error)).
						Do(
							codegen.Expr(`
fieldValues := ` + file.Use("github.com/go-courier/sqlx/builder", "FieldValuesFromStructByNonZero") + `(m, zeroFields...)
return m.` + methodForUpdateWithMap + `(db, fieldValues)
`),
						),
				)
			}

			{

				method := createMethod("FetchBy%sForUpdate", fieldNamesWithoutEnabled...)

				file.WriteBlock(
					codegen.Func(codegen.Var(
						codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx", "DB"))), "db")).
						Named(method).
						MethodOf(codegen.Var(m.PtrType(), "m")).
						Return(codegen.Var(codegen.Error)).
						Do(
							m.SnippetEnableIfNeed(file),

							codegen.Expr(`
table := m.T()

err := db.QueryExprAndScan(
`+file.Use("github.com/go-courier/sqlx/builder", "Select")+`(nil).
From(
	m.T(),
`+file.Use("github.com/go-courier/sqlx/builder", "Where")+`(`+toExactlyConditionFrom(file, fieldNames...)+`),
`+file.Use("github.com/go-courier/sqlx/builder", "ForUpdate")+`(),
`+file.Use("github.com/go-courier/sqlx/builder", "Comment")+`(?),
),
m,
)
`,
								file.Val(m.StructName+"."+method),
							),

							codegen.Return(codegen.Expr("err")),
						),
				)
			}

			{
				methodForDelete := createMethod("DeleteBy%s", fieldNamesWithoutEnabled...)

				file.WriteBlock(
					codegen.Func(codegen.Var(
						codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx", "DB"))), "db")).
						Named(methodForDelete).
						MethodOf(codegen.Var(m.PtrType(), "m")).
						Return(codegen.Var(codegen.Error)).
						Do(
							m.SnippetEnableIfNeed(file),

							codegen.Expr(`
table := m.T()

_, err := db.ExecExpr(
`+file.Use("github.com/go-courier/sqlx/builder", "Delete")+`().
From(
	m.T(),
`+file.Use("github.com/go-courier/sqlx/builder", "Where")+`(`+toExactlyConditionFrom(file, fieldNames...)+`),
`+file.Use("github.com/go-courier/sqlx/builder", "Comment")+`(?),
),
)
`,
								file.Val(m.StructName+"."+methodForDelete),
							),

							codegen.Return(codegen.Expr("err")),
						),
				)

				if m.HasSoftDelete {

					methodForSoftDelete := createMethod("SoftDeleteBy%s", fieldNamesWithoutEnabled...)

					file.WriteBlock(
						codegen.Func(codegen.Var(
							codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx", "DB"))), "db")).
							Named(methodForSoftDelete).
							MethodOf(codegen.Var(m.PtrType(), "m")).
							Return(codegen.Var(codegen.Error)).
							Do(
								m.SnippetEnableIfNeed(file),

								codegen.Expr(`
table := m.T()

fieldValues := `+file.Use("github.com/go-courier/sqlx/builder", "FieldValues")+`{}`),

								m.SnippetDisableIfNeedForFieldValues(file),
								m.SnippetSetUpdatedAtIfNeedForFieldValues(file),

								codegen.Expr(`

_, err := db.ExecExpr(
`+file.Use("github.com/go-courier/sqlx/builder", "Update")+`(m.T()).
	Where(
		`+toExactlyConditionFrom(file, fieldNames...)+`,
		`+file.Use("github.com/go-courier/sqlx/builder", "Comment")+`(?),
	).
	Set(table.AssignmentsByFieldValues(fieldValues)...),
)

if err != nil {
	dbErr := `+file.Use("github.com/go-courier/sqlx", "DBErr")+`(err)
	if dbErr.IsConflict() {
		return 	m.`+methodForDelete+`(db)
	}
}

return nil
`,

									file.Val(m.StructName+"."+methodForSoftDelete),
								),
							),
					)
				}
			}
		}
	})
}

func (m *Model) WriteCRUD(file *codegen.File) {
	m.WriteCreate(file)
	m.WriteDelete(file)
	m.WriteByKey(file)
}

func toExactlyConditionFrom(file *codegen.File, fieldNames ...string) string {
	buf := &bytes.Buffer{}
	buf.WriteString(file.Use("github.com/go-courier/sqlx/builder", "And"))
	buf.WriteString(`(
`)

	for _, fieldName := range fieldNames {
		buf.WriteString(fmt.Sprintf(`table.F("%s").Eq(m.%s),
		`, fieldName, fieldName))
	}

	buf.WriteString(`
)`)

	return buf.String()
}

func createMethod(method string, fieldNames ...string) string {
	return fmt.Sprintf(method, strings.Join(fieldNames, "And"))
}
func (m *Model) FieldType(file *codegen.File, fieldName string) string {
	if field, ok := m.Fields[fieldName]; ok {
		typ := field.Type().String()
		if strings.Index(typ, ".") > -1 {
			return file.Use(packagesx.GetPkgImportPathAndExpose(typ))
		}
		return typ
	}
	return ""
}
