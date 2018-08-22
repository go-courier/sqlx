package datatypes

import (
	"fmt"
	"strings"
)

type DataType string

func (dataType DataType) Is(dataTypes []DataType) bool {
	for _, d := range dataTypes {
		if dataType.String() == d.String() {
			return true
		}
	}
	return false
}

func (dataType DataType) String() string {
	return strings.ToLower(string(dataType))
}

func (dataType DataType) WithArgs(args ...string) string {
	return fmt.Sprintf("%s(%s)", dataType.String(), strings.Join(args, ","))
}

func (dataType DataType) IsInteger() bool {
	return dataType.Is([]DataType{
		"INTEGER",
		"INT",
		"SMALLINT",
		"TINYINT",
		"MEDIUMINT",
		"BIGINT",
	})
}

func (dataType DataType) IsFloating() bool {
	return dataType.Is([]DataType{
		"FLOAT",
		"DOUBLE",
	})
}

func (dataType DataType) IsFixed() bool {
	return dataType.Is([]DataType{
		"DECIMAL",
		"NUMERIC",
	})
}

func (dataType DataType) IsChar() bool {
	return dataType.Is([]DataType{
		"CHAR",
		"VARCHAR",
		"BINARY",
		"VARBINARY",
	})
}

func (dataType DataType) IsText() bool {
	return dataType.Is([]DataType{
		"TINYTEXT",
		"MEDIUMTEXT",
		"TEXT",
		"LONGTEXT",
	})
}

func (dataType DataType) IsEnum() bool {
	return dataType.Is([]DataType{
		"ENUM",
	})
}

func (dataType DataType) IsSet() bool {
	return dataType.Is([]DataType{
		"SET",
	})
}

func (dataType DataType) IsBlob() bool {
	return dataType.Is([]DataType{
		"TINYBLOB",
		"MEDIUMBLOB",
		"BLOB",
		"LONGBLOB",
	})
}

func (dataType DataType) IsDate() bool {
	return dataType.Is([]DataType{
		"DATE",
		"TIME",
		"DATETIME",
		"TIMESTAMP",
		"YEAR",
	})
}
