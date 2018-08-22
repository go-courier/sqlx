package nullable

import (
	"database/sql"
	_ "unsafe"
)

func NewNullIgnoreScanner(dest interface{}) *NullIgnoreScanner {
	return &NullIgnoreScanner{
		dest: dest,
	}
}

type NullIgnoreScanner struct {
	dest interface{}
}

func (scanner *NullIgnoreScanner) Scan(src interface{}) error {
	if scanner, ok := scanner.dest.(sql.Scanner); ok {
		return scanner.Scan(src)
	}
	if src == nil {
		return nil
	}
	return convertAssign(scanner.dest, src)
}

//go:linkname convertAssign database/sql.convertAssign
func convertAssign(dest, src interface{}) error
