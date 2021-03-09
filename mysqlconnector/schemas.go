package mysqlconnector

import (
	"database/sql"
	"strings"

	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/builder"
)

func toInterfaces(list ...string) []interface{} {
	s := make([]interface{}, len(list))
	for i, v := range list {
		s[i] = v
	}
	return s
}

func dbFromInformationSchema(db sqlx.DBExecutor) (*sqlx.Database, error) {
	d := db.D()
	tableNames := d.Tables.TableNames()

	database := sqlx.NewDatabase(d.Name)

	tableColumnSchema := SchemaDatabase.T(&ColumnSchema{})
	columnSchemaList := make([]ColumnSchema, 0)

	err := db.QueryExprAndScan(
		builder.Select(tableColumnSchema.Columns.Clone()).
			From(tableColumnSchema,
				builder.Where(
					builder.And(
						tableColumnSchema.F("TABLE_SCHEMA").Eq(database.Name),
						tableColumnSchema.F("TABLE_NAME").In(toInterfaces(tableNames...)...),
					),
				),
			),
		&columnSchemaList,
	)
	if err != nil {
		return nil, err
	}

	for i := range columnSchemaList {
		columnSchema := columnSchemaList[i]
		table := database.Table(columnSchema.TABLE_NAME)
		if table == nil {
			table = builder.T(columnSchema.TABLE_NAME)
			database.AddTable(table)
		}
		table.AddCol(colFromColumnSchema(&columnSchema))
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
							tableIndexSchema.F("TABLE_SCHEMA").Eq(database.Name),
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
			return nil, err
		}

		for _, indexSchema := range indexList {
			table := database.Table(indexSchema.TABLE_NAME)

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

	return database, nil
}

var SchemaDatabase = sqlx.NewDatabase("INFORMATION_SCHEMA")

func init() {
	SchemaDatabase.Register(&ColumnSchema{})
	SchemaDatabase.Register(&IndexSchema{})
}

func colFromColumnSchema(columnSchema *ColumnSchema) *builder.Column {
	col := builder.Col(columnSchema.COLUMN_NAME)

	col.AutoIncrement = strings.Contains(columnSchema.EXTRA, "auto_increment")

	defaultValue := columnSchema.COLUMN_DEFAULT

	if defaultValue.Valid {
		v := normalizeDefaultValue(defaultValue.String)
		col.Default = &v
	}

	if strings.Contains(columnSchema.EXTRA, "on update ") {
		v := strings.Split(columnSchema.EXTRA, "on update ")[1]
		col.OnUpdate = &v
	}

	dataType := columnSchema.DATA_TYPE

	if strings.HasSuffix(columnSchema.COLUMN_TYPE, "unsigned") {
		dataType = dataType + " unsigned"
	}

	col.GetDataType = func(engine string) string {
		return dataType
	}

	// numeric type
	if columnSchema.NUMERIC_PRECISION > 0 {
		col.Length = columnSchema.NUMERIC_PRECISION
		col.Decimal = columnSchema.NUMERIC_SCALE
	} else {
		col.Length = columnSchema.CHARACTER_MAXIMUM_LENGTH
	}

	if columnSchema.IS_NULLABLE == "YES" {
		col.Null = true
	}

	return col
}

// https://dev.mysql.com/doc/refman/8.0/en/data-type-defaults.html
func normalizeDefaultValue(v string) string {
	if len(v) == 0 {
		return "''"
	}

	switch v {
	case "CURRENT_TIMESTAMP", "CURRENT_DATE", "NULL":
		return v
	}

	// functions
	if strings.Contains(v, "(") && strings.Contains(v, ")") {
		return v
	}

	return quoteWith(v, '\'', false, false)
}

type ColumnSchema struct {
	TABLE_SCHEMA             string         `db:"TABLE_SCHEMA"`
	TABLE_NAME               string         `db:"TABLE_NAME"`
	COLUMN_NAME              string         `db:"COLUMN_NAME"`
	DATA_TYPE                string         `db:"DATA_TYPE"`
	COLUMN_TYPE              string         `db:"COLUMN_TYPE"`
	EXTRA                    string         `db:"EXTRA"`
	IS_NULLABLE              string         `db:"IS_NULLABLE"`
	COLUMN_DEFAULT           sql.NullString `db:"COLUMN_DEFAULT"`
	CHARACTER_MAXIMUM_LENGTH uint64         `db:"CHARACTER_MAXIMUM_LENGTH"`
	NUMERIC_PRECISION        uint64         `db:"NUMERIC_PRECISION"`
	NUMERIC_SCALE            uint64         `db:"NUMERIC_SCALE"`
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
