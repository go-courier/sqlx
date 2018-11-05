package generator

import (
	"fmt"

	"github.com/go-courier/codegen"
	"github.com/go-courier/packagesx"
)

func (m *Model) WriteCount(file *codegen.File) {
	file.WriteBlock(
		codegen.Func(
			codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2", "DB"))), "db"),
			codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2/builder", "Condition"))), "condition"),
			codegen.Var(codegen.Ellipsis(codegen.Type(file.Use("github.com/go-courier/sqlx/v2/builder", "Addition"))), "additions"),
		).
			Named("Count").
			MethodOf(codegen.Var(m.PtrType(), "m")).
			Return(
				codegen.Var(codegen.Int),
				codegen.Var(codegen.Error),
			).
			Do(
				codegen.Expr(`
count := -1

table := m.T()
_ = table
`),

				func() codegen.Snippet {
					if m.HasSoftDelete {
						return codegen.Expr(
							`condition = ?(condition, table.F("`+m.FieldKeySoftDelete+`").Eq(?))`,
							codegen.Id(file.Use("github.com/go-courier/sqlx/v2/builder", "And")),
							codegen.Id(file.Use(packagesx.GetPkgImportPathAndExpose(m.ConstSoftDeleteTrue))),
						)
					}
					return nil
				}(),

				codegen.Expr(`

finalAdditions := []`+file.Use("github.com/go-courier/sqlx/v2/builder", "Addition")+`{
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Where")+`(condition),
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Comment")+`(?),
}

if len(additions) > 0 {
	finalAdditions = append(finalAdditions, additions...)
}

err := db.QueryExprAndScan(
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Select")+`(
	`+file.Use("github.com/go-courier/sqlx/v2/builder", "Count")+`(),
).
From(m.T(), finalAdditions...),
&count,
)

return count, err
`,
					file.Val(m.StructName+".Count"),
				),
			),
	)
}

func (m *Model) WriteList(file *codegen.File) {
	file.WriteBlock(
		codegen.Func(
			codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2", "DB"))), "db"),
			codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2/builder", "Condition"))), "condition"),
			codegen.Var(codegen.Ellipsis(codegen.Type(file.Use("github.com/go-courier/sqlx/v2/builder", "Addition"))), "additions"),
		).
			Named("List").
			MethodOf(codegen.Var(m.PtrType(), "m")).
			Return(
				codegen.Var(codegen.Slice(codegen.Type(m.StructName))),
				codegen.Var(codegen.Error),
			).
			Do(
				codegen.Expr(`
list := make([]`+m.StructName+`, 0)

table := m.T()
_ = table
`),

				func() codegen.Snippet {
					if m.HasSoftDelete {
						return codegen.Expr(
							`condition = ?(condition, table.F("`+m.FieldKeySoftDelete+`").Eq(?))`,
							codegen.Id(file.Use("github.com/go-courier/sqlx/v2/builder", "And")),
							codegen.Id(file.Use(packagesx.GetPkgImportPathAndExpose(m.ConstSoftDeleteTrue))),
						)
					}
					return nil
				}(),

				codegen.Expr(`

finalAdditions := []`+file.Use("github.com/go-courier/sqlx/v2/builder", "Addition")+`{
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Where")+`(condition),
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Comment")+`(?),
}

if len(additions) > 0 {
	finalAdditions = append(finalAdditions, additions...)
}

err := db.QueryExprAndScan(
`+file.Use("github.com/go-courier/sqlx/v2/builder", "Select")+`(nil).
From(m.T(), finalAdditions...),
&list,
)

return list, err
`,
					file.Val(m.StructName+".List"),
				),
			),
	)
}

func (m *Model) WriteBatchList(file *codegen.File) {
	indexedFields := m.IndexFieldNames()

	for _, field := range indexedFields {
		method := fmt.Sprintf("BatchFetchBy%sList", field)

		typ := m.FieldType(file, field)

		file.WriteBlock(
			codegen.Func(
				codegen.Var(codegen.Star(codegen.Type(file.Use("github.com/go-courier/sqlx/v2", "DB"))), "db"),
				codegen.Var(codegen.Slice(typ), "values"),
			).
				Named(method).
				MethodOf(codegen.Var(m.PtrType(), "m")).
				Return(
					codegen.Var(codegen.Slice(codegen.Type(m.StructName))),
					codegen.Var(codegen.Error),
				).
				Do(
					codegen.Expr(`
if len(values) == 0 {
	return nil, nil
}

table := m.T()

condition := table.F("` + field + `").In(values)

return m.List(db, condition)
`),
				),
		)
	}
}
