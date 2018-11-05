package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-courier/codegen"
	"github.com/go-courier/packagesx"
	"github.com/go-courier/sqlx/v2/builder"
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
			m.FieldType(file, m.FieldKeyCreatedAt),
			codegen.Id(file.Use("time", "Now")))
	}

	return nil
}

func (m *Model) SnippetSetLastInsertIdIfNeed(file *codegen.File) codegen.Snippet {
	if m.HasAutoIncrement {
		return codegen.Expr(`
if err == nil {
	lastInsertID, _ := result.LastInsertId()
	m.? = ?(lastInsertID)
}
`,
			codegen.Id(m.FieldKeyAutoIncrement),
			m.FieldType(file, m.FieldKeyAutoIncrement),
		)
	}

	return codegen.Expr(`
_ = result
`)
}

func (m *Model) SnippetSetUpdatedAtIfNeed(file *codegen.File) codegen.Snippet {
	if m.HasUpdatedAt {
		return codegen.Expr(`
if m.?.IsZero() {
	m.? = ?(?())
}
`,
			codegen.Id(m.FieldKeyUpdatedAt),
			codegen.Id(m.FieldKeyUpdatedAt),
			m.FieldType(file, m.FieldKeyUpdatedAt),
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
			m.FieldType(file, m.FieldKeyUpdatedAt),
			codegen.Id(file.Use("time", "Now")))
	}
	return nil
}

func (m *Model) WriteCreate(file *codegen.File) {
	file.WriteBlock(
		codegen.Func(codegen.Var(
			codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2", "DB"))), "db")).
			Named("Create").
			MethodOf(codegen.Var(m.PtrType(), "m")).
			Return(codegen.Var(codegen.Error)).
			Do(
				m.SnippetEnableIfNeed(file),
				m.SnippetSetCreatedAtIfNeed(file),
				m.SnippetSetUpdatedAtIfNeed(file),

				codegen.Expr(`
d := m.D()

switch db.DriverName() {
	case "mysql":
		result, err := db.ExecExpr(d.Insert(m, nil))
		?
		return err
	case "postgres":
		return db.QueryExprAndScan(d.Insert(m, nil, `+file.Use("github.com/go-courier/sqlx/v2/builder", "Returning")+`(nil)), m)
}

return nil
`,
					m.SnippetSetLastInsertIdIfNeed(file),
				),
			),
	)

	if len(m.Keys.UniqueIndexes) > 0 {

		file.WriteBlock(
			codegen.Func(
				codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2", "DB"))), "db"),
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
					m.SnippetSetUpdatedAtIfNeed(file),

					codegen.Expr(`
fieldValues := `+file.Use("github.com/go-courier/sqlx/v2/builder", "FieldValuesFromStructByNonZero")+`(m, updateFields...)
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

switch db.DriverName() {
	case "mysql":
		_, err := db.ExecExpr(`+file.Use("github.com/go-courier/sqlx/v2/builder", "Insert")+`().
			Into(
				table,
				`+file.Use("github.com/go-courier/sqlx/v2/builder", "OnDuplicateKeyUpdate")+`(table.AssignmentsByFieldValues(fieldValues)...),
				`+file.Use("github.com/go-courier/sqlx/v2/builder", "Comment")+`(?),
			).
			Values(cols, vals...),
		)
		return err
	case "postgres":
		indexes := m.UniqueIndexes()
		fields := make([]string, 0)
		for _, fs := range indexes {
			fields = append(fields, fs...)
		}
		indexFields, _ := m.T().Fields(fields...)

		_, err := db.ExecExpr(github_com_go_courier_sqlx_builder.Insert().
			Into(
				table,
				`+file.Use("github.com/go-courier/sqlx/v2/builder", "OnConflict")+`(indexFields).
					DoUpdateSet(table.AssignmentsByFieldValues(fieldValues)...),
				`+file.Use("github.com/go-courier/sqlx/v2/builder", "Comment")+`(?),
			).
			Values(cols, vals...),
		)
		return err
	}
return nil
`,
						file.Val(m.StructName+".CreateOnDuplicateWithUpdateFields"),
						file.Val(m.StructName+".CreateOnDuplicateWithUpdateFields"),
					),
				),
		)
	}

}

func (m *Model) WriteDelete(file *codegen.File) {
	file.WriteBlock(
		codegen.Func(codegen.Var(
			codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2", "DB"))), "db")).
			Named("DeleteByStruct").
			MethodOf(codegen.Var(m.PtrType(), "m")).
			Return(codegen.Var(codegen.Error)).
			Do(
				m.SnippetEnableIfNeed(file),
				m.SnippetSetUpdatedAtIfNeed(file),

				codegen.Expr(`
_, err := db.ExecExpr(
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Delete")+`().
From(
	m.T(),
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Where")+`(m.ConditionByStruct()),
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Comment")+`(?),
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

		if m.HasSoftDelete && key.IsPrimary() {
			fieldNames = append(fieldNames, m.FieldKeySoftDelete)
		}

		if key.IsUnique {
			{
				methodForFetch := createMethod("FetchBy%s", fieldNamesWithoutEnabled...)

				file.WriteBlock(
					codegen.Func(codegen.Var(
						codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2", "DB"))), "db")).
						Named(methodForFetch).
						MethodOf(codegen.Var(m.PtrType(), "m")).
						Return(codegen.Var(codegen.Error)).
						Do(
							m.SnippetEnableIfNeed(file),

							codegen.Expr(`
table := m.T()

err := db.QueryExprAndScan(
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Select")+`(nil).
From(
	m.T(),
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Where")+`(`+toExactlyConditionFrom(file, fieldNames...)+`),
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Comment")+`(?),
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
						codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2", "DB"))), "db"),
						codegen.Var(codegen.Type(file.Use("github.com/go-courier/sqlx/v2/builder", "FieldValues")), "fieldValues"),
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
	`+file.Use("github.com/go-courier/sqlx/v2/builder", "Update")+`(m.T()).
		Where(
			`+toExactlyConditionFrom(file, fieldNames...)+`,
			`+file.Use("github.com/go-courier/sqlx/v2/builder", "Comment")+`(?),
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
						codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2", "DB"))), "db"),
						codegen.Var(codegen.Ellipsis(codegen.String), "zeroFields"),
					).
						Named(methodForUpdateWithStruct).
						MethodOf(codegen.Var(m.PtrType(), "m")).
						Return(codegen.Var(codegen.Error)).
						Do(
							codegen.Expr(`
fieldValues := ` + file.Use("github.com/go-courier/sqlx/v2/builder", "FieldValuesFromStructByNonZero") + `(m, zeroFields...)
return m.` + methodForUpdateWithMap + `(db, fieldValues)
`),
						),
				)
			}

			{

				method := createMethod("FetchBy%sForUpdate", fieldNamesWithoutEnabled...)

				file.WriteBlock(
					codegen.Func(codegen.Var(
						codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2", "DB"))), "db")).
						Named(method).
						MethodOf(codegen.Var(m.PtrType(), "m")).
						Return(codegen.Var(codegen.Error)).
						Do(
							m.SnippetEnableIfNeed(file),

							codegen.Expr(`
table := m.T()

err := db.QueryExprAndScan(
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Select")+`(nil).
From(
	m.T(),
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Where")+`(`+toExactlyConditionFrom(file, fieldNames...)+`),
`+file.Use("github.com/go-courier/sqlx/v2/builder", "ForUpdate")+`(),
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Comment")+`(?),
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
						codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2", "DB"))), "db")).
						Named(methodForDelete).
						MethodOf(codegen.Var(m.PtrType(), "m")).
						Return(codegen.Var(codegen.Error)).
						Do(
							m.SnippetEnableIfNeed(file),

							codegen.Expr(`
table := m.T()

_, err := db.ExecExpr(
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Delete")+`().
From(
	m.T(),
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Where")+`(`+toExactlyConditionFrom(file, fieldNames...)+`),
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Comment")+`(?),
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
							codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2", "DB"))), "db")).
							Named(methodForSoftDelete).
							MethodOf(codegen.Var(m.PtrType(), "m")).
							Return(codegen.Var(codegen.Error)).
							Do(
								m.SnippetEnableIfNeed(file),

								codegen.Expr(`
table := m.T()

fieldValues := `+file.Use("github.com/go-courier/sqlx/v2/builder", "FieldValues")+`{}`),

								m.SnippetDisableIfNeedForFieldValues(file),
								m.SnippetSetUpdatedAtIfNeedForFieldValues(file),

								codegen.Expr(`

_, err := db.ExecExpr(
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Update")+`(m.T()).
	Where(
		`+toExactlyConditionFrom(file, fieldNames...)+`,
		`+file.Use("github.com/go-courier/sqlx/v2/builder", "Comment")+`(?),
	).
	Set(table.AssignmentsByFieldValues(fieldValues)...),
)

if err != nil {
	dbErr := `+file.Use("github.com/go-courier/sqlx/v2", "DBErr")+`(err)
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
	buf.WriteString(file.Use("github.com/go-courier/sqlx/v2/builder", "And"))
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

func (m *Model) FieldType(file *codegen.File, fieldName string) codegen.SnippetType {
	if field, ok := m.Fields[fieldName]; ok {
		typ := field.Type().String()
		if strings.Index(typ, ".") > -1 {
			importPath, name := packagesx.GetPkgImportPathAndExpose(typ)
			if importPath != m.TypeName.Pkg().Path() {
				return codegen.Type(file.Use(importPath, name))
			}
			return codegen.Type(name)
		}
		return codegen.BuiltInType(typ)
	}
	return nil
}
