package mysqlconnector

import _ "unsafe"

//go:linkname quoteWith strconv.quoteWith
func quoteWith(s string, quote byte, ASCIIonly, graphicOnly bool) string
