package datatypes

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-courier/enumeration"
)

var (
	reColumnType = regexp.MustCompile("(\\w+)(\\((\\d+)(,(\\d+))?\\))?")
	reDefault    = regexp.MustCompile("DEFAULT (('([^']+)?')|(\\w+)(\\((\\d+)(,(\\d+))?\\))?)")
	reCharset    = regexp.MustCompile("CHARACTER SET (\\S+)")
	reCollate    = regexp.MustCompile("COLLATE (\\S+)")
)

func pickDefault(s string) (sql string, defaultValue string, exists bool, err error) {
	sql = s
	values := reDefault.FindStringSubmatch(sql)
	if len(values) == 9 {
		exists = true
		sql = reDefault.ReplaceAllString(sql, "")
		if values[4] != "" {
			defaultValue = values[4]
			if !ValueKeywords[defaultValue] {
				err = fmt.Errorf("unsupport default value %s", defaultValue)
				return
			}
		} else {
			defaultValue = values[3]
		}
	}
	return
}

func ParseColumnType(sql string) (columnType *ColumnType, err error) {
	columnType = &ColumnType{}

	sql, columnType.Default, columnType.HasDefault, err = pickDefault(sql)
	if err != nil {
		return
	}

	dataTypeAndDesc := reColumnType.FindStringSubmatch(sql)
	if len(dataTypeAndDesc) != 6 {
		err = fmt.Errorf("missing sql data type")
		return
	}

	columnType.DataType = DataType(strings.ToLower(dataTypeAndDesc[1]))
	if dataTypeAndDesc[3] != "" {
		columnType.Length, _ = strconv.ParseInt(dataTypeAndDesc[3], 10, 64)
	}
	if dataTypeAndDesc[5] != "" {
		columnType.Decimal, _ = strconv.ParseInt(dataTypeAndDesc[5], 10, 64)
	}

	upperSQL := strings.ToUpper(sql)

	columnType.NotNull = strings.Contains(upperSQL, "NOT NULL")

	if columnType.IsInteger() || columnType.IsFloating() {
		columnType.AutoIncrement = strings.Contains(upperSQL, "AUTO_INCREMENT")
	}

	if columnType.IsInteger() || columnType.IsFloating() || columnType.IsFixed() {
		columnType.Zerofill = strings.Contains(upperSQL, "ZEROFILL")
		columnType.Unsigned = strings.Contains(upperSQL, "UNSIGNED") || columnType.Zerofill
	}

	if columnType.IsChar() || columnType.IsText() {
		charset := "utf8"
		{
			values := reCharset.FindStringSubmatch(upperSQL)
			if len(values) == 2 {
				columnType.Charset = strings.ToLower(values[1])
				charset = columnType.Charset
			}
		}
		if strings.Contains(upperSQL, "BINARY") {
			columnType.Collate = charset + "_bin"
		}
		{
			values := reCollate.FindStringSubmatch(upperSQL)
			if len(values) == 2 {
				columnType.Collate = strings.ToLower(values[1])
			}
		}
	}

	if columnType.IsDate() {
		columnType.OnUpdateByCurrentTimestamp = strings.Contains(upperSQL, "ON UPDATE CURRENT_TIMESTAMP")
	}

	columnType.Zerofill = strings.Contains(upperSQL, "ZEROFILL")

	return
}

type ColumnType struct {
	// type
	DataType
	// length
	Length int64
	// 小数点位数
	Decimal int64
	// todo enum 值 or set
	Values   []string
	Unsigned bool
	// [CHARACTER SET utf8 when text] COLLATE utf8_bin
	Charset string
	Collate string

	Zerofill bool
	NotNull  bool

	Enum enumeration.Enum

	HasDefault bool
	Default    string

	// extra
	AutoIncrement              bool
	OnUpdateByCurrentTimestamp bool

	Comment string
}

func (columnType ColumnType) IsEnum() bool {
	return columnType.Enum != nil
}

func (columnType ColumnType) DeAlias() *ColumnType {
	if columnType.Charset == "" {
		columnType.Charset = "utf8"
	}

	switch columnType.DataType.String() {
	case "timestamp", "datetime":
		if columnType.HasDefault && columnType.Default == "0" {
			columnType.Default = "0000-00-00 00:00:00"
			if columnType.Length > 0 {
				columnType.Default += "." + strings.Repeat("0", int(columnType.Length))
			}
		}
		return &columnType
	case "boolean":
		columnType.DataType = "tinyint"
		columnType.Length = 1
		return &columnType
	default:
		if columnType.Length == 0 {
			switch columnType.DataType.String() {
			case "tinyint":
				columnType.Length = 3
			case "smallint":
				columnType.Length = 6
			case "int":
				columnType.Length = 11
			case "bigint":
				columnType.Length = 20
			}
		}
		return &columnType
	}
}

func (columnType *ColumnType) asDefaultValue() string {
	if ValueKeywords[strings.ToUpper(columnType.Default)] {
		if columnType.IsDate() && columnType.Length > 0 {
			return fmt.Sprintf("%s(%d)", columnType.Default, columnType.Length)
		}
		return columnType.Default
	}
	return `'` + columnType.Default + `'`
}

func (columnType ColumnType) String() string {
	b := bytes.Buffer{}
	if len(columnType.Values) > 0 {
		b.WriteString(columnType.DataType.WithArgs(columnType.Values...))
	} else if columnType.Length > 0 {
		if columnType.Decimal > 0 {
			b.WriteString(columnType.DataType.WithArgs(
				fmt.Sprintf("%d", columnType.Length),
				fmt.Sprintf("%d", columnType.Decimal),
			))
		} else {
			b.WriteString(columnType.DataType.WithArgs(
				fmt.Sprintf("%d", columnType.Length),
			))
		}
	} else {
		b.WriteString(columnType.DataType.String())
	}

	if columnType.Unsigned {
		b.WriteString(" unsigned")
	}

	if columnType.Zerofill {
		b.WriteString(" zerofill")
	}

	{
		charset := "utf8"

		if columnType.Charset != "" {
			b.WriteString(" CHARACTER SET " + columnType.Charset)
			charset = columnType.Charset
		}

		if columnType.Collate != "" {
			if columnType.Collate == charset+"_bin" {
				b.WriteString(" binary")
			} else {
				b.WriteString(" COLLATE " + columnType.Collate)
			}
		}
	}

	if columnType.NotNull {
		b.WriteString(" NOT NULL")
	}

	if columnType.AutoIncrement {
		b.WriteString(" AUTO_INCREMENT")
	} else {
		if columnType.HasDefault {
			b.WriteString(fmt.Sprintf(" DEFAULT %s", columnType.asDefaultValue()))
		}
		if columnType.IsDate() && columnType.OnUpdateByCurrentTimestamp {
			if columnType.Length > 0 {
				b.WriteString(fmt.Sprintf(" ON UPDATE CURRENT_TIMESTAMP(%d)", columnType.Length))
			} else {
				b.WriteString(" ON UPDATE CURRENT_TIMESTAMP")
			}
		}
	}

	if columnType.Comment != "" {
		b.WriteString(fmt.Sprintf(" COMMENT '%s'", strings.Replace(columnType.Comment, "'", "\\'", -1)))
	}

	return b.String()
}

var ValueKeywords = map[string]bool{
	"NULL":              true,
	"CURRENT_TIMESTAMP": true,
	"CURRENT_DATE":      true,
}
