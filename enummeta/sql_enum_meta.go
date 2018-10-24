package enummeta

import (
	"github.com/go-courier/sqlx/builder"
)

type SqlMetaEnum struct {
	TName string `db:"F_table_name" sql:"varchar(64) NOT NULL"`
	CName string `db:"F_column_name" sql:"varchar(64) NOT NULL"`
	Value int    `db:"F_value" sql:"int NOT NULL"`
	Type  string `db:"F_type" sql:"varchar(255) NOT NULL"`
	Key   string `db:"F_key"  sql:"varchar(255) NOT NULL"`
	Label string `db:"F_label" sql:"varchar(255) NOT NULL"`
}

func (*SqlMetaEnum) TableName() string {
	return "t_sql_meta_enum"
}

func (*SqlMetaEnum) UniqueIndexes() builder.Indexes {
	return builder.Indexes{"I_enum": builder.FieldNames{"TName", "CName", "Value"}}
}
