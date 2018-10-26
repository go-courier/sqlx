package enummeta

import (
	"github.com/go-courier/sqlx/builder"
)

type SqlMetaEnum struct {
	TName string `db:"F_table_name,size=64"`
	CName string `db:"F_column_name,size=64"`
	Value int    `db:"F_value"`
	Type  string `db:"F_type,size=255"`
	Key   string `db:"F_key,size=255"`
	Label string `db:"F_label,size=255"`
}

func (*SqlMetaEnum) TableName() string {
	return "t_sql_meta_enum"
}

func (*SqlMetaEnum) UniqueIndexes() builder.Indexes {
	return builder.Indexes{"I_enum": {"TName", "CName", "Value"}}
}
