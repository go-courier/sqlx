package postgresqlconnector

import (
	"regexp"
	"strings"

	"github.com/go-courier/sqlx/v2/builder"

	"github.com/go-courier/sqlx/v2"
)

func toInterfaces(list ...string) []interface{} {
	s := make([]interface{}, len(list))
	for i, v := range list {
		s[i] = v
	}
	return s
}

func dbFromInformationSchema(db *sqlx.DB) *sqlx.Database {
	dbName := db.Name
	dbSchema := db.Schema
	tableNames := db.Tables.TableNames()

	d := sqlx.NewDatabase(dbName).WithSchema(dbSchema)

	tableColumnSchema := SchemaDatabase.T(&ColumnSchema{}).WithSchema("information_schema")
	columnSchemaList := make([]ColumnSchema, 0)

	tableSchema := "public"
	if d.Schema != "" {
		tableSchema = d.Schema
	}

	stmt := builder.Select(tableColumnSchema.Columns.Clone()).From(tableColumnSchema,
		builder.Where(
			builder.And(
				tableColumnSchema.F("TABLE_SCHEMA").Eq(tableSchema),
				tableColumnSchema.F("TABLE_NAME").In(toInterfaces(tableNames...)...),
			),
		),
	)

	err := db.QueryExprAndScan(stmt, &columnSchemaList)
	if err != nil {
		panic(err)
	}

	for _, columnSchema := range columnSchemaList {
		table := d.Table(columnSchema.TABLE_NAME)
		if table == nil {
			table = builder.T(columnSchema.TABLE_NAME)
			d.AddTable(table)
		}
		col := builder.Col(columnSchema.COLUMN_NAME)
		table.AddCol(col)
	}

	if tableColumnSchema.Columns.Len() != 0 {
		tableIndexSchema := SchemaDatabase.T(&IndexSchema{})

		indexList := make([]IndexSchema, 0)

		err = db.QueryExprAndScan(
			builder.Select(tableIndexSchema.Columns.Clone()).
				From(
					tableIndexSchema,
					builder.Where(
						builder.And(
							tableIndexSchema.F("TABLE_SCHEMA").Eq(tableSchema),
							tableIndexSchema.F("TABLE_NAME").In(toInterfaces(tableNames...)...),
						),
					),
				),
			&indexList,
		)

		if err != nil {
			panic(err)
		}

		for _, indexSchema := range indexList {
			table := d.Table(indexSchema.TABLE_NAME)

			key := &builder.Key{}
			key.Name = indexSchema.INDEX_NAME[len(table.Name)+1:]
			key.Method = strings.ToUpper(regexp.MustCompile(`USING ([^ ]+)`).FindString(indexSchema.INDEX_DEF)[6:])
			key.IsUnique = strings.Index(indexSchema.INDEX_DEF, "UNIQUE") > -1

			fields := regexp.MustCompile(`\([^\)]+\)`).FindString(indexSchema.INDEX_DEF)
			if len(fields) > 0 {
				fields = fields[1 : len(fields)-1]
			}
			key.Columns, _ = table.Cols(strings.Split(fields, ", ")...)
			table.AddKey(key)
		}
	}

	return d
}

var SchemaDatabase = sqlx.NewDatabase("INFORMATION_SCHEMA")

func init() {
	SchemaDatabase.Register(&ColumnSchema{})
	SchemaDatabase.Register(&IndexSchema{})
}

type ColumnSchema struct {
	TABLE_SCHEMA string `db:"table_schema"`
	TABLE_NAME   string `db:"table_name"`
	COLUMN_NAME  string `db:"column_name"`
}

func (ColumnSchema) TableName() string {
	return "columns"
}

type IndexSchema struct {
	TABLE_SCHEMA string `db:"schemaname"`
	TABLE_NAME   string `db:"tablename"`
	INDEX_NAME   string `db:"indexname"`
	INDEX_DEF    string `db:"indexdef"`
}

func (IndexSchema) TableName() string {
	return "pg_indexes"
}
