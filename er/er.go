package er

import (
	"reflect"
	"strings"

	"github.com/go-courier/enumeration"
	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/builder"
)

func DatabaseERFromDB(database *sqlx.Database, dialect builder.Dialect) *ERDatabase {
	erd := &ERDatabase{Name: database.Name, Tables: map[string]*ERTable{}}

	database.Tables.Range(func(table *builder.Table, idx int) {
		t := &ERTable{Name: table.Name, Cols: map[string]*ERCol{}, Keys: map[string]*ERKey{}}

		if len(table.Description) > 0 {
			t.Summary = table.Description[0]

			if len(table.Description) > 1 {
				t.Desc = strings.Join(table.Description[1:], "\n")
			}
		}

		erd.Tables[t.Name] = t

		table.Columns.Range(func(col *builder.Column, idx int) {
			if col.DeprecatedActions != nil {
				return
			}

			c := &ERCol{
				Name:     col.Name,
				DataType: dialect.DataType(col.ColumnType).Expr().String(),
			}

			if len(col.Description) > 0 {
				c.Summary = col.Description[0]

				if len(col.Description) > 1 {
					c.Desc = strings.Join(col.Description[1:], "\n")
				}
			}

			rv := reflect.New(col.ColumnType.Type)

			if rv.CanInterface() {
				if emum, ok := rv.Interface().(enumeration.Enum); ok {

					c.Enum = map[string]EREnum{}

					for _, e := range emum.ConstValues() {
						em := EREnum{}
						em.Value = e.Int()
						em.Name = e.String()
						em.Label = e.Label()

						c.Enum[em.Name] = em
					}
				}
			}

			if len(col.Relation) == 2 {
				relTable := database.Tables.Model(col.Relation[0])
				if relTable != nil && relTable != table {
					relCol := relTable.F(col.Relation[1])
					if relCol != nil {
						c.Rel = []string{
							relTable.Name,
							relCol.Name,
						}
					}
				}
			}

			t.Cols[c.Name] = c
		})

		table.Keys.Range(func(key *builder.Key, idx int) {
			k := &ERKey{
				Name:      key.Name,
				Method:    key.Method,
				IsUnique:  key.IsUnique,
				IsPrimary: key.Name == "primary",
			}

			key.Columns.Range(func(col *builder.Column, idx int) {
				k.Cols = append(k.Cols, col.Name)
			})

			t.Keys[k.Name] = k
		})
	})

	return erd
}

type ERDatabase struct {
	Name   string              `json:"name"`
	Tables map[string]*ERTable `json:"tables"`
}

type ERTable struct {
	Name    string            `json:"name"`
	Summary string            `json:"summary"`
	Desc    string            `json:"desc"`
	Cols    map[string]*ERCol `json:"cols"`
	Keys    map[string]*ERKey `json:"keys"`
}

type ERCol struct {
	Name     string            `json:"name"`
	DataType string            `json:"dataType"`
	Enum     map[string]EREnum `json:"enum"`
	Summary  string            `json:"summary"`
	Desc     string            `json:"desc"`
	Rel      []string          `json:"rel"`
}

type EREnum struct {
	Value int    `json:"value"`
	Name  string `json:"name"`
	Label string `json:"label"`
}

type ERKey struct {
	Name      string   `json:"name"`
	Method    string   `json:"method"`
	IsUnique  bool     `json:"isUnique"`
	IsPrimary bool     `json:"isPrimary"`
	Cols      []string `json:"cols"`
}
