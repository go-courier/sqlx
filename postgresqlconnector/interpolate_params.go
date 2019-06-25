package postgresqlconnector

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"strconv"
)

func interpolateParams(query string, args []driver.NamedValue) fmt.Stringer {

	return &SqlPrinter{
		query: query,
		args:  args,
	}
}

type SqlPrinter struct {
	query string
	args  []driver.NamedValue
}

func (p *SqlPrinter) String() string {
	buf := bytes.NewBufferString(p.query)

	buf.WriteString(" | ")

	for i, a := range p.args {
		if i > 0 {
			buf.WriteString(", ")
		}
		if bs, ok := a.Value.([]byte); ok {
			buf.WriteByte('[')
			for i, b := range bs {
				if i != 0 {
					buf.WriteByte(',')
				}
				if i > 8 {
					buf.WriteString("...")
					break
				}
				buf.WriteString(strconv.FormatUint(uint64(b), 10))
			}

			buf.WriteByte(']')
		} else {
			buf.WriteString(fmt.Sprintf("%v", a.Value))
		}
	}

	return buf.String()
}
