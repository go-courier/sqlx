package datatypes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultParse(t *testing.T) {
	tt := require.New(t)
	{
		results := reDefault.FindStringSubmatch("DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP")
		tt.Equal("CURRENT_TIMESTAMP", results[1])
	}

	{
		results := reDefault.FindStringSubmatch("DEFAULT CURRENT_TIMESTAMP(5)")
		tt.Equal("CURRENT_TIMESTAMP(5)", results[1])
	}

	{
		results := reDefault.FindStringSubmatch("DEFAULT '1' ON UPDATE '1'")
		tt.Equal("1", results[3])
	}

	{
		results := reDefault.FindStringSubmatch("DEFAULT '1 2'")
		tt.Equal("1 2", results[3])
	}

	{
		results := reDefault.FindStringSubmatch("DEFAULT NULL")
		tt.Equal("NULL", results[1])
	}
}

func TestColumnType_unsigned_not_null_auto_increment(t *testing.T) {
	RunTestCase(t, &TestCase{
		sql:      "bigint(64) unsigned NOT NULL AUTO_INCREMENT",
		finalSql: "bigint(64) unsigned NOT NULL AUTO_INCREMENT",
		ColumnType: &ColumnType{
			DataType:      "bigint",
			Length:        64,
			Unsigned:      true,
			NotNull:       true,
			AutoIncrement: true,
		},
	})
}

func TestColumnType_with_charset(t *testing.T) {
	RunTestCase(t, &TestCase{
		sql:      "varchar(64) CHARACTER SET latin1 NOT NULL",
		finalSql: "varchar(64) CHARACTER SET latin1 NOT NULL",
		ColumnType: &ColumnType{
			DataType:   "varchar",
			Length:     64,
			NotNull:    true,
			HasDefault: false,
			Charset:    "latin1",
		},
	})
}

func TestColumnType_with_charset_binary(t *testing.T) {
	RunTestCase(t, &TestCase{
		sql:      "varchar(64) CHARACTER SET latin1 COLLATE latin1_bin NOT NULL",
		finalSql: "varchar(64) CHARACTER SET latin1 binary NOT NULL",
		ColumnType: &ColumnType{
			DataType:   "varchar",
			Length:     64,
			NotNull:    true,
			HasDefault: false,
			Charset:    "latin1",
			Collate:    "latin1_bin",
		},
	})
}

func TestColumnType_with_charset_and_collate(t *testing.T) {
	RunTestCase(t, &TestCase{
		sql:      "varchar(64) CHARACTER SET latin1 COLLATE latin1_german1_ci NOT NULL",
		finalSql: "varchar(64) CHARACTER SET latin1 COLLATE latin1_german1_ci NOT NULL",
		ColumnType: &ColumnType{
			DataType:   "varchar",
			Length:     64,
			NotNull:    true,
			HasDefault: false,
			Charset:    "latin1",
			Collate:    "latin1_german1_ci",
		},
	})
}

func TestColumnType_drop_which_cannot_auto_increment(t *testing.T) {
	RunTestCase(t, &TestCase{
		sql:      "varchar(64) unsigned NOT NULL AUTO_INCREMENT",
		finalSql: "varchar(64) NOT NULL",
		ColumnType: &ColumnType{
			DataType:   "varchar",
			Length:     64,
			NotNull:    true,
			HasDefault: false,
		},
	})
}

func TestColumnType_varchar_binary_default_null(t *testing.T) {
	RunTestCase(t, &TestCase{
		sql:      "varchar(64) binary DEFAULT NULL",
		finalSql: "varchar(64) binary DEFAULT NULL",
		ColumnType: &ColumnType{
			DataType:   "varchar",
			Collate:    "utf8_bin",
			Length:     64,
			HasDefault: true,
			Default:    "NULL",
		},
	})
}

func TestColumnType_bigint_unsigned_not_null_default_1(t *testing.T) {
	RunTestCase(t, &TestCase{
		sql:      "bigint(64) unsigned NOT NULL DEFAULT '1'",
		finalSql: "bigint(64) unsigned NOT NULL DEFAULT '1'",
		ColumnType: &ColumnType{
			DataType:   "bigint",
			Length:     64,
			Unsigned:   true,
			NotNull:    true,
			HasDefault: true,
			Default:    "1",
		},
	})
}

func TestColumnType_datetime(t *testing.T) {
	RunTestCase(t, &TestCase{
		sql:      "datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP",
		finalSql: "datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP",
		ColumnType: &ColumnType{
			DataType:                   "datetime",
			NotNull:                    true,
			HasDefault:                 true,
			Default:                    "CURRENT_TIMESTAMP",
			OnUpdateByCurrentTimestamp: true,
		},
	})
}

func TestColumnType_datetime_with_len(t *testing.T) {
	RunTestCase(t, &TestCase{
		sql:      "datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP",
		finalSql: "datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)",
		ColumnType: &ColumnType{
			DataType:                   "datetime",
			Length:                     6,
			NotNull:                    true,
			HasDefault:                 true,
			Default:                    "CURRENT_TIMESTAMP",
			OnUpdateByCurrentTimestamp: true,
		},
	})
}

func TestColumnType_datetime_zero(t *testing.T) {
	RunTestCase(t, &TestCase{
		sql:      "datetime NOT NULL DEFAULT '0' ON UPDATE CURRENT_TIMESTAMP",
		finalSql: "datetime NOT NULL DEFAULT '0' ON UPDATE CURRENT_TIMESTAMP",
		ColumnType: &ColumnType{
			DataType:                   "datetime",
			NotNull:                    true,
			HasDefault:                 true,
			Default:                    "0",
			OnUpdateByCurrentTimestamp: true,
		},
	})
}

type TestCase struct {
	sql        string
	finalSql   string
	ColumnType *ColumnType
}

func RunTestCase(t *testing.T, c *TestCase) {
	tt := require.New(t)
	ct, err := ParseColumnType(c.sql)
	tt.NoError(err)
	tt.Equal(c.ColumnType, ct, c.sql)
	tt.Equal(c.finalSql, c.ColumnType.String(), c.finalSql)
}
