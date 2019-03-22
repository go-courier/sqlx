package generator

import (
	"go/types"
	"strings"

	"github.com/go-courier/codegen"
	"github.com/go-courier/packagesx"
	"github.com/go-courier/sqlx/v2/builder"
)

func NewModel(pkg *packagesx.Package, typeName *types.TypeName, comments string, cfg *Config) *Model {
	m := Model{}
	m.Config = cfg
	m.Config.SetDefaults()

	m.TypeName = typeName

	m.Table = builder.T(cfg.TableName)

	p := pkg.Pkg(typeName.Pkg().Path())

	forEachStructField(typeName.Type().Underlying().Underlying().(*types.Struct), func(structVal *types.Var, columnName string, tagValue string) {
		col := builder.Col(columnName).Field(structVal.Name()).Type("", tagValue)

		for id, o := range p.TypesInfo.Defs {
			if o == structVal {
				doc := pkg.CommentsOf(id)

				rel, lines := parseColRelFromComment(doc)

				if rel != "" {
					relPath := strings.Split(rel, ".")

					if len(relPath) != 2 {
						continue
					}

					col.Relation = relPath
				}

				if len(lines) > 0 {
					col.Comment = lines[0]
					col.Description = lines
				}
			}
		}

		m.addColumn(col, structVal)
	})

	m.HasDeletedAt = m.Table.F(m.FieldKeyDeletedAt) != nil
	m.HasCreatedAt = m.Table.F(m.FieldKeyCreatedAt) != nil
	m.HasUpdatedAt = m.Table.F(m.FieldKeyUpdatedAt) != nil

	keys, lines := parseKeysFromDoc(comments)
	m.Keys = keys

	if len(lines) > 0 {
		m.Description = lines
	}

	if m.HasDeletedAt {
		m.Keys.PatchUniqueIndexesWithSoftDelete(m.FieldKeyDeletedAt)
	}
	m.Keys.Bind(m.Table)

	if autoIncrementCol := m.Table.AutoIncrement(); autoIncrementCol != nil {
		m.HasAutoIncrement = true
		m.FieldKeyAutoIncrement = autoIncrementCol.FieldName
	}

	return &m
}

type Model struct {
	*types.TypeName
	*Config
	*Keys
	*builder.Table
	Fields                map[string]*types.Var
	FieldKeyAutoIncrement string
	HasDeletedAt          bool
	HasCreatedAt          bool
	HasUpdatedAt          bool
	HasAutoIncrement      bool
}

func (m *Model) addColumn(col *builder.Column, tpe *types.Var) {
	m.Table.Columns.Add(col)
	if m.Fields == nil {
		m.Fields = map[string]*types.Var{}
	}
	m.Fields[col.FieldName] = tpe
}

func (m *Model) WriteTo(file *codegen.File) {
	m.WriteTableKeyInterfaces(file)

	if m.WithTableInterfaces {
		m.WriteTableInterfaces(file)
	}

	if m.WithMethods {
		m.WriteCRUD(file)
		m.WriteList(file)
		m.WriteCount(file)
		m.WriteBatchList(file)
	}
}

func (m *Model) Type() codegen.SnippetType {
	return codegen.Type(m.StructName)
}

func (m *Model) PtrType() codegen.SnippetType {
	return codegen.Star(m.Type())
}

func (m *Model) VarTable() string {
	return m.StructName + "Table"
}
