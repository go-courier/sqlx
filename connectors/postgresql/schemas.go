package postgresql

import (
	"fmt"
	"regexp"
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

	dbName := d.Name
	dbSchema := d.Schema
	tableNames := d.Tables.TableNames()

	d = sqlx.NewDatabase(dbName).WithSchema(dbSchema)

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
		return nil, err
	}

	for i := range columnSchemaList {
		columnSchema := columnSchemaList[i]

		table := d.Table(columnSchema.TABLE_NAME)
		if table == nil {
			table = builder.T(columnSchema.TABLE_NAME)
			d.AddTable(table)
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
							tableIndexSchema.F("TABLE_SCHEMA").Eq(tableSchema),
							tableIndexSchema.F("TABLE_NAME").In(toInterfaces(tableNames...)...),
						),
					),
				),
			&indexList,
		)

		if err != nil {
			return nil, err
		}

		for _, indexSchema := range indexList {
			table := d.Table(indexSchema.TABLE_NAME)

			key := &builder.Key{}
			key.Name = strings.ToLower(indexSchema.INDEX_NAME[len(table.Name)+1:])
			key.Method = strings.ToUpper(regexp.MustCompile(`USING ([^ ]+)`).FindString(indexSchema.INDEX_DEF)[6:])
			key.IsUnique = strings.Contains(indexSchema.INDEX_DEF, "UNIQUE")

			fields := regexp.MustCompile(`\([^\)]+\)`).FindString(indexSchema.INDEX_DEF)
			if len(fields) > 0 {
				fields = fields[1 : len(fields)-1]
			}
			key.Columns, _ = table.Cols(strings.Split(fields, ", ")...)
			table.AddKey(key)
		}
	}

	return d, nil
}

var SchemaDatabase = sqlx.NewDatabase("INFORMATION_SCHEMA")

func init() {
	SchemaDatabase.Register(&ColumnSchema{})
	SchemaDatabase.Register(&IndexSchema{})
}

func colFromColumnSchema(columnSchema *ColumnSchema) *builder.Column {
	col := builder.Col(columnSchema.COLUMN_NAME)

	defaultValue := columnSchema.COLUMN_DEFAULT

	if defaultValue != "" {
		col.AutoIncrement = strings.HasSuffix(columnSchema.COLUMN_DEFAULT, "_seq'::regclass)")

		if !col.AutoIncrement {
			if !strings.Contains(defaultValue, "'::") && '0' <= defaultValue[0] && defaultValue[0] <= '9' {
				defaultValue = fmt.Sprintf("'%s'::integer", defaultValue)
			}
			col.Default = &defaultValue
		}
	}

	dataType := columnSchema.DATA_TYPE

	if col.AutoIncrement {
		if strings.HasPrefix(dataType, "big") {
			dataType = "bigserial"
		} else {
			dataType = "serial"
		}
	}

	col.DataType = dataType

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

type ColumnSchema struct {
	TABLE_SCHEMA             string `db:"table_schema"`
	TABLE_NAME               string `db:"table_name"`
	COLUMN_NAME              string `db:"column_name"`
	DATA_TYPE                string `db:"data_type"`
	IS_NULLABLE              string `db:"is_nullable"`
	COLUMN_DEFAULT           string `db:"column_default"`
	CHARACTER_MAXIMUM_LENGTH uint64 `db:"character_maximum_length"`
	NUMERIC_PRECISION        uint64 `db:"numeric_precision"`
	NUMERIC_SCALE            uint64 `db:"numeric_scale"`
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
