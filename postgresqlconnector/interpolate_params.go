package postgresqlconnector

import (
	"bytes"
	"database/sql/driver"
	"fmt"
)

func interpolateParams(query string, args []driver.Value) (string, error) {
	buf := bytes.NewBufferString(query)

	buf.WriteString(" | ")

	for i, a := range args {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%v", a))
	}

	return buf.String(), nil
}
