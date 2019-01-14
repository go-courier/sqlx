package postgresqlconnector

import (
	"bytes"
	"database/sql/driver"
	"fmt"
)

func interpolateParams(query string, args []driver.NamedValue) string {
	buf := bytes.NewBufferString(query)

	buf.WriteString(" | ")

	for i, a := range args {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%v", a.Value))
	}

	return buf.String()
}
