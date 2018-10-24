package mysql

import (
	"database/sql"
	"github.com/go-courier/sqlx"
	"strings"
	"time"

	"github.com/go-courier/sqlx/builder"
	"github.com/go-courier/sqlx/datatypes"
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
				tableColumnSchema.F("TABLE_SCHEMA").Eq(d.Name).
					And(tableColumnSchema.F("TABLE_NAME").In(toInterfaces(tableNames...)...)),
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
			table = builder.T(d.Database, columnSchema.TABLE_NAME)
			d.Database.Register(table)
		}
		col := builder.Col(table, columnSchema.COLUMN_NAME)
		col.ColumnType = *columnSchema.ToColumnType()
		table.Columns.Add(col)
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

			key, exists := table.Keys.Key(indexSchema.INDEX_NAME)
			if !exists {
				switch indexSchema.INDEX_TYPE {
				case "SPATIAL":
					key = builder.SpatialIndex(indexSchema.INDEX_NAME)
				default:
					if indexSchema.INDEX_NAME == string(builder.PRIMARY) {
						key = builder.PrimaryKey()
					} else if indexSchema.NON_UNIQUE == 1 {
						key = builder.Index(indexSchema.INDEX_NAME)
					} else {
						key = builder.UniqueIndex(indexSchema.INDEX_NAME)
					}
				}
				table.Keys.Add(key)
			}

			key.WithCols(table.Col(indexSchema.COLUMN_NAME))
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

func (columnSchema ColumnSchema) ToColumnType() *datatypes.ColumnType {
	columnType, _ := datatypes.ParseColumnType(columnSchema.COLUMN_TYPE)
	columnType.NotNull = columnSchema.IS_NULLABLE == "NO"
	if columnSchema.EXTRA == "auto_increment" {
		columnType.AutoIncrement = true
	}
	if strings.HasPrefix(columnSchema.EXTRA, "on update CURRENT_TIMESTAMP") {
		columnType.OnUpdateByCurrentTimestamp = true
	}

	if columnSchema.CHARACTER_SET_NAME.Valid {
		columnType.Charset = columnSchema.CHARACTER_SET_NAME.String
	}

	if columnSchema.COLLATION_NAME.Valid {
		if collationDefaults[columnType.Charset] != columnSchema.COLLATION_NAME.String {
			columnType.Collate = columnSchema.COLLATION_NAME.String
		}
	}

	if columnSchema.COLUMN_DEFAULT.Valid {
		columnType.HasDefault = true
		columnType.Default = columnSchema.COLUMN_DEFAULT.String
	} else if columnSchema.IS_NULLABLE == "YES" {
		columnType.HasDefault = true
		columnType.Default = "NULL"
	}

	columnType.Comment = columnSchema.COLUMN_COMMENT

	return columnType
}

var collationDefaults = map[string]string{
	"utf8mb4":  "utf8mb4_general_ci",
	"utf8":     "utf8_general_ci",
	"utf32":    "utf32_general_ci",
	"utf16le":  "utf16le_general_ci",
	"utf16":    "utf16_general_ci",
	"ujis":     "ujis_japanese_ci",
	"ucs2":     "ucs2_general_ci",
	"tis620":   "tis620_thai_ci",
	"swe7":     "swe7_swedish_ci",
	"sjis":     "sjis_japanese_ci",
	"macroman": "macroman_general_ci",
	"macce":    "macce_general_ci",
	"latin7":   "latin7_general_ci",
	"latin5":   "latin5_turkish_ci",
	"latin2":   "latin2_general_ci",
	"latin1":   "latin1_swedish_ci",
	"koi8u":    "koi8u_general_ci",
	"koi8r":    "koi8r_general_ci",
	"keybcs2":  "keybcs2_general_ci",
	"hp8":      "hp8_english_ci",
	"hebrew":   "hebrew_general_ci",
	"greek":    "greek_general_ci",
	"geostd8":  "geostd8_general_ci",
	"gbk":      "gbk_chinese_ci",
	"gb2312":   "gb2312_chinese_ci",
	"euckr":    "euckr_korean_ci",
	"eucjpms":  "eucjpms_japanese_ci",
	"dec8":     "dec8_swedish_ci",
	"cp932":    "cp932_japanese_ci",
	"cp866":    "cp866_general_ci",
	"cp852":    "cp852_general_ci",
	"cp850":    "cp850_general_ci",
	"cp1257":   "cp1257_general_ci",
	"cp1256":   "cp1256_general_ci",
	"cp1251":   "cp1251_general_ci",
	"cp1250":   "cp1250_general_ci",
	"binary":   "binary",
	"big5":     "big5_chinese_ci",
	"ascii":    "ascii_general_ci",
	"armscii8": "armscii8_general_ci",
}

type Schema struct {
	SCHEMA_NAME string `db:"SCHEMA_NAME" sql:"varchar(64) NOT NULL DEFAULT ''"`
}

func (schema Schema) TableName() string {
	return "SCHEMATA"
}

type TableSchema struct {
	TABLE_SCHEMA    string         `db:"TABLE_SCHEMA" sql:"varchar(64) NOT NULL DEFAULT ''"`
	TABLE_NAME      string         `db:"TABLE_NAME" sql:"varchar(64) NOT NULL DEFAULT ''"`
	TABLE_TYPE      string         `db:"TABLE_TYPE" sql:"varchar(64) NOT NULL DEFAULT ''"`
	ENGINE          sql.NullString `db:"ENGINE" sql:"varchar(64) DEFAULT NULL"`
	VERSION         sql.NullInt64  `db:"VERSION" sql:"bigint(21) unsigned DEFAULT NULL"`
	ROW_FORMAT      sql.NullString `db:"ROW_FORMAT" sql:"varchar(10) DEFAULT NULL"`
	TABLE_ROWS      sql.NullInt64  `db:"TABLE_ROWS" sql:"bigint(21) unsigned DEFAULT NULL"`
	AVG_ROW_LENGTH  sql.NullInt64  `db:"AVG_ROW_LENGTH" sql:"bigint(21) unsigned DEFAULT NULL"`
	DATA_LENGTH     sql.NullInt64  `db:"DATA_LENGTH" sql:"bigint(21) unsigned DEFAULT NULL"`
	MAX_DATA_LENGTH sql.NullInt64  `db:"MAX_DATA_LENGTH" sql:"bigint(21) unsigned DEFAULT NULL"`
	INDEX_LENGTH    sql.NullInt64  `db:"INDEX_LENGTH" sql:"bigint(21) unsigned DEFAULT NULL"`
	DATA_FREE       sql.NullInt64  `db:"DATA_FREE" sql:"bigint(21) unsigned DEFAULT NULL"`
	AUTO_INCREMENT  sql.NullString `db:"AUTO_INCREMENT" sql:"bigint(21) unsigned DEFAULT NULL"`
	CREATE_TIME     time.Time      `db:"CREATE_TIME" sql:"datetime DEFAULT NULL"`
	UPDATE_TIME     time.Time      `db:"UPDATE_TIME" sql:"datetime DEFAULT NULL"`
	CHECK_TIME      time.Time      `db:"CHECK_TIME" sql:"datetime DEFAULT NULL"`
	TABLE_COLLATION sql.NullString `db:"TABLE_COLLATION" sql:"varchar(32) DEFAULT NULL"`
	CHECKSUM        sql.NullInt64  `db:"CHECKSUM" sql:"bigint(21) unsigned DEFAULT NULL"`
	CREATE_OPTIONS  sql.NullString `db:"CREATE_OPTIONS" sql:"varchar(255) DEFAULT NULL"`
	TABLE_COMMENT   string         `db:"TABLE_COMMENT" sql:"varchar(2048) NOT NULL DEFAULT ''"`
}

func (tableSchema TableSchema) TableName() string {
	return "TABLES"
}

type ColumnSchema struct {
	TABLE_SCHEMA             string         `db:"TABLE_SCHEMA" sql:"varchar(64) NOT NULL DEFAULT ''"`
	TABLE_NAME               string         `db:"TABLE_NAME" sql:"varchar(64) NOT NULL DEFAULT ''"`
	COLUMN_NAME              string         `db:"COLUMN_NAME" sql:"varchar(64) NOT NULL DEFAULT ''"`
	ORDINAL_POSITION         int32          `db:"ORDINAL_POSITION" sql:"bigint(21) unsigned NOT NULL DEFAULT '0'"`
	COLUMN_DEFAULT           sql.NullString `db:"COLUMN_DEFAULT" sql:"longtext"`
	IS_NULLABLE              string         `db:"IS_NULLABLE" sql:"varchar(3) NOT NULL DEFAULT ''"`
	DATA_TYPE                string         `db:"DATA_TYPE" sql:"varchar(64) NOT NULL DEFAULT ''"`
	CHARACTER_MAXIMUM_LENGTH sql.NullInt64  `db:"CHARACTER_MAXIMUM_LENGTH" sql:"bigint(21) unsigned DEFAULT NULL"`
	CHARACTER_OCTET_LENGTH   sql.NullInt64  `db:"CHARACTER_OCTET_LENGTH" sql:"bigint(21) unsigned DEFAULT NULL"`
	NUMERIC_PRECISION        sql.NullInt64  `db:"NUMERIC_PRECISION" sql:"bigint(21) unsigned DEFAULT NULL"`
	NUMERIC_SCALE            sql.NullInt64  `db:"NUMERIC_SCALE" sql:"bigint(21) unsigned DEFAULT NULL"`
	DATETIME_PRECISION       sql.NullInt64  `db:"DATETIME_PRECISION" sql:"bigint(21) unsigned DEFAULT NULL"`
	CHARACTER_SET_NAME       sql.NullString `db:"CHARACTER_SET_NAME" sql:"varchar(32) DEFAULT NULL"`
	COLLATION_NAME           sql.NullString `db:"COLLATION_NAME" sql:"varchar(32) DEFAULT NULL"`
	COLUMN_TYPE              string         `db:"COLUMN_TYPE" sql:"longtext NOT NULL"`
	COLUMN_KEY               string         `db:"COLUMN_KEY" sql:"varchar(3) NOT NULL DEFAULT ''"`
	EXTRA                    string         `db:"EXTRA" sql:"varchar(30) NOT NULL DEFAULT ''"`
	PRIVILEGES               string         `db:"PRIVILEGES" sql:"varchar(80) NOT NULL DEFAULT ''"`
	COLUMN_COMMENT           string         `db:"COLUMN_COMMENT" sql:"varchar(1024) NOT NULL DEFAULT ''"`
}

func (columnSchema ColumnSchema) TableName() string {
	return "COLUMNS"
}

type IndexSchema struct {
	TABLE_SCHEMA  string         `db:"TABLE_SCHEMA" sql:"varchar(64) NOT NULL DEFAULT ''"`
	TABLE_NAME    string         `db:"TABLE_NAME" sql:"varchar(64) NOT NULL DEFAULT ''"`
	NON_UNIQUE    int32          `db:"NON_UNIQUE" sql:"bigint(1) NOT NULL DEFAULT '0'"`
	INDEX_SCHEMA  string         `db:"INDEX_SCHEMA" sql:"varchar(64) NOT NULL DEFAULT ''"`
	INDEX_NAME    string         `db:"INDEX_NAME" sql:"varchar(64) NOT NULL DEFAULT ''"`
	SEQ_IN_INDEX  int32          `db:"SEQ_IN_INDEX" sql:"bigint(2) NOT NULL DEFAULT '0'"`
	COLUMN_NAME   string         `db:"COLUMN_NAME" sql:"varchar(64) NOT NULL DEFAULT ''"`
	COLLATION     sql.NullString `db:"COLLATION" sql:"varchar(1) DEFAULT NULL"`
	CARDINALITY   sql.NullInt64  `db:"CARDINALITY" sql:"bigint(21) DEFAULT NULL"`
	SUB_PART      sql.NullInt64  `db:"SUB_PART" sql:"bigint(3) DEFAULT NULL"`
	PACKED        sql.NullString `db:"PACKED" sql:"varchar(10) DEFAULT NULL"`
	NULLABLE      string         `db:"NULLABLE" sql:"varchar(3) NOT NULL DEFAULT ''"`
	INDEX_TYPE    string         `db:"INDEX_TYPE" sql:"varchar(16) NOT NULL DEFAULT ''"`
	COMMENT       string         `db:"COMMENT" sql:"varchar(16) DEFAULT NULL"`
	INDEX_COMMENT string         `db:"INDEX_COMMENT" sql:"varchar(1024) NOT NULL DEFAULT ''"`
}

func (indexSchema IndexSchema) TableName() string {
	return "STATISTICS"
}
