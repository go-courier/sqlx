package postgresql

import (
	"bytes"
	"sort"
	"strings"
)

func FromConfigString(s string) PostgreSQLOpts {
	opts := PostgreSQLOpts{}
	for _, kv := range strings.Split(s, " ") {
		kvs := strings.Split(kv, "=")
		if len(kvs) > 1 {
			opts[kvs[0]] = kvs[1]
		}
	}
	return opts
}

type PostgreSQLOpts map[string]string

func (opts PostgreSQLOpts) String() string {
	buf := bytes.NewBuffer(nil)

	kvs := make([]string, 0)
	for k := range opts {
		kvs = append(kvs, k)
	}
	sort.Strings(kvs)

	for i, k := range kvs {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(opts[k])
	}

	return buf.String()
}
