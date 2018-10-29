package mysqlconnector

import (
	"database/sql"
	"github.com/go-courier/sqlx"
	"time"

	"github.com/go-courier/sqlx/builder"
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

	schema := SchemaDatabase.T(&Schema{})

	schemaList := make([]Schema, 0)
	errForSchema := db.QueryExprAndScan(
		builder.Select(nil).From(
			schema,
			builder.Where(schema.F("SCHEMA_NAME").Eq(d.Name)),
			builder.Limit(1),
		),
		&schemaList,
	)
	if errForSchema != nil {
		return nil
	}
	if len(schemaList) == 0 {
		return nil
	}

	tableColumnSchema := SchemaDatabase.T(&ColumnSchema{})
	columnSchemaList := make([]ColumnSchema, 0)

	err := db.QueryExprAndScan(
		builder.Select(nil).From(tableColumnSchema,
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
			builder.Select(nil).
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
	SchemaDatabase.Register(&Schema{})
	SchemaDatabase.Register(&TableSchema{})
	SchemaDatabase.Register(&ColumnSchema{})
	SchemaDatabase.Register(&IndexSchema{})
}

type Schema struct {
	SCHEMA_NAME string `db:"SCHEMA_NAME"`
}

func (Schema) TableName() string {
	return "INFORMATION_SCHEMA.SCHEMATA"
}

type TableSchema struct {
	TABLE_SCHEMA    string         `db:"TABLE_SCHEMA"`
	TABLE_NAME      string         `db:"TABLE_NAME"`
	TABLE_TYPE      string         `db:"TABLE_TYPE"`
	ENGINE          sql.NullString `db:"ENGINE"`
	VERSION         sql.NullInt64  `db:"VERSION"`
	ROW_FORMAT      sql.NullString `db:"ROW_FORMAT"`
	TABLE_ROWS      sql.NullInt64  `db:"TABLE_ROWS"`
	AVG_ROW_LENGTH  sql.NullInt64  `db:"AVG_ROW_LENGTH"`
	DATA_LENGTH     sql.NullInt64  `db:"DATA_LENGTH"`
	MAX_DATA_LENGTH sql.NullInt64  `db:"MAX_DATA_LENGTH"`
	INDEX_LENGTH    sql.NullInt64  `db:"INDEX_LENGTH"`
	DATA_FREE       sql.NullInt64  `db:"DATA_FREE"`
	AUTO_INCREMENT  sql.NullString `db:"AUTO_INCREMENT"`
	CREATE_TIME     time.Time      `db:"CREATE_TIME"`
	UPDATE_TIME     time.Time      `db:"UPDATE_TIME"`
	CHECK_TIME      time.Time      `db:"CHECK_TIME"`
	TABLE_COLLATION sql.NullString `db:"TABLE_COLLATION"`
	CHECKSUM        sql.NullInt64  `db:"CHECKSUM"`
	CREATE_OPTIONS  sql.NullString `db:"CREATE_OPTIONS"`
	TABLE_COMMENT   string         `db:"TABLE_COMMENT"`
}

func (TableSchema) TableName() string {
	return "INFORMATION_SCHEMA.TABLES"
}

type ColumnSchema struct {
	TABLE_SCHEMA             string         `db:"TABLE_SCHEMA"`
	TABLE_NAME               string         `db:"TABLE_NAME"`
	COLUMN_NAME              string         `db:"COLUMN_NAME"`
	ORDINAL_POSITION         int32          `db:"ORDINAL_POSITION"`
	COLUMN_DEFAULT           sql.NullString `db:"COLUMN_DEFAULT"`
	IS_NULLABLE              string         `db:"IS_NULLABLE"`
	DATA_TYPE                string         `db:"DATA_TYPE"`
	CHARACTER_MAXIMUM_LENGTH sql.NullInt64  `db:"CHARACTER_MAXIMUM_LENGTH"`
	CHARACTER_OCTET_LENGTH   sql.NullInt64  `db:"CHARACTER_OCTET_LENGTH"`
	NUMERIC_PRECISION        sql.NullInt64  `db:"NUMERIC_PRECISION"`
	NUMERIC_SCALE            sql.NullInt64  `db:"NUMERIC_SCALE"`
	DATETIME_PRECISION       sql.NullInt64  `db:"DATETIME_PRECISION"`
	CHARACTER_SET_NAME       sql.NullString `db:"CHARACTER_SET_NAME"`
	COLLATION_NAME           sql.NullString `db:"COLLATION_NAME"`
	COLUMN_TYPE              string         `db:"COLUMN_TYPE"`
	COLUMN_KEY               string         `db:"COLUMN_KEY"`
	EXTRA                    string         `db:"EXTRA"`
	PRIVILEGES               string         `db:"PRIVILEGES"`
	COLUMN_COMMENT           string         `db:"COLUMN_COMMENT"`
}

func (ColumnSchema) TableName() string {
	return "INFORMATION_SCHEMA.COLUMNS"
}

type IndexSchema struct {
	TABLE_SCHEMA  string         `db:"TABLE_SCHEMA"`
	TABLE_NAME    string         `db:"TABLE_NAME"`
	NON_UNIQUE    int32          `db:"NON_UNIQUE"`
	INDEX_SCHEMA  string         `db:"INDEX_SCHEMA"`
	INDEX_NAME    string         `db:"INDEX_NAME"`
	SEQ_IN_INDEX  int32          `db:"SEQ_IN_INDEX"`
	COLUMN_NAME   string         `db:"COLUMN_NAME"`
	COLLATION     sql.NullString `db:"COLLATION"`
	CARDINALITY   sql.NullInt64  `db:"CARDINALITY"`
	SUB_PART      sql.NullInt64  `db:"SUB_PART"`
	PACKED        sql.NullString `db:"PACKED"`
	NULLABLE      string         `db:"NULLABLE"`
	INDEX_TYPE    string         `db:"INDEX_TYPE"`
	COMMENT       string         `db:"COMMENT"`
	INDEX_COMMENT string         `db:"INDEX_COMMENT"`
}

func (IndexSchema) TableName() string {
	return "INFORMATION_SCHEMA.STATISTICS"
}
