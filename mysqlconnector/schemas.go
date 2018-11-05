package mysqlconnector

import (
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

func DBFromInformationSchema(db *sqlx.DB, dbName string, tableNames ...string) *sqlx.Database {
	d := sqlx.NewDatabase(dbName)

	tableColumnSchema := SchemaDatabase.T(&ColumnSchema{})
	columnSchemaList := make([]ColumnSchema, 0)

	err := db.QueryExprAndScan(
		builder.Select(tableColumnSchema.Columns.Clone()).From(tableColumnSchema,
			builder.Where(
				builder.And(
					tableColumnSchema.F("TABLE_SCHEMA").Eq(d.Name),
					tableColumnSchema.F("TABLE_NAME").In(toInterfaces(tableNames...)...),
				),
			),
		),
		&columnSchemaList,
	)
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
							tableIndexSchema.F("TABLE_SCHEMA").Eq(d.Name),
							tableIndexSchema.F("TABLE_NAME").In(toInterfaces(tableNames...)...),
						),
					),
					builder.OrderBy(
						builder.AscOrder(tableIndexSchema.F("INDEX_NAME")),
						builder.AscOrder(tableIndexSchema.F("SEQ_IN_INDEX")),
					),
				),
			&indexList,
		)

		if err != nil {
			panic(err)
		}

		for _, indexSchema := range indexList {
			table := d.Table(indexSchema.TABLE_NAME)

			if key := table.Keys.Key(indexSchema.INDEX_NAME); key != nil {
				key.Columns.Add(table.Col(indexSchema.COLUMN_NAME))
			} else {
				key := &builder.Key{}
				key.Name = indexSchema.INDEX_NAME
				key.Method = indexSchema.INDEX_TYPE
				key.IsUnique = indexSchema.NON_UNIQUE == 0
				key.Columns, _ = table.Cols(indexSchema.COLUMN_NAME)
				table.AddKey(key)
			}
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
	TABLE_SCHEMA string `db:"TABLE_SCHEMA"`
	TABLE_NAME   string `db:"TABLE_NAME"`
	COLUMN_NAME  string `db:"COLUMN_NAME"`
}

func (ColumnSchema) TableName() string {
	return "INFORMATION_SCHEMA.COLUMNS"
}

type IndexSchema struct {
	TABLE_SCHEMA string `db:"TABLE_SCHEMA"`
	TABLE_NAME   string `db:"TABLE_NAME"`
	NON_UNIQUE   int32  `db:"NON_UNIQUE"`
	INDEX_NAME   string `db:"INDEX_NAME"`
	SEQ_IN_INDEX int32  `db:"SEQ_IN_INDEX"`
	COLUMN_NAME  string `db:"COLUMN_NAME"`
	INDEX_TYPE   string `db:"INDEX_TYPE"`
}

func (IndexSchema) TableName() string {
	return "INFORMATION_SCHEMA.STATISTICS"
}
