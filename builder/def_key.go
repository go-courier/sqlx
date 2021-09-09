package builder

import (
	"strings"
)

func PrimaryKey(columns *Columns) *Key {
	return UniqueIndex("PRIMARY", columns)
}

func UniqueIndex(name string, columns *Columns, exprs ...string) *Key {
	key := Index(name, columns, exprs...)
	key.IsUnique = true
	return key
}

func Index(name string, columns *Columns, exprs ...string) *Key {
	k := &Key{
		Name: strings.ToLower(name),
	}

	if columns != nil {
		k.Def.FieldNames = columns.FieldNames()
		k.Def.ColNames = columns.ColNames()
	}

	if len(exprs) > 0 {
		k.Def.Expr = strings.Join(exprs, " ")
	}

	return k
}

var _ TableDefinition = (*Key)(nil)

func ParseIndexDef(parts ...string) *IndexDef {
	fe := IndexDef{}

	if len(parts) == 1 {
		s := parts[0]

		if strings.Contains(s, "#") || strings.Contains(s, "(") {
			fe.Expr = s
		} else {
			fe.FieldNames = strings.Split(s, " ")
		}
	} else {
		fe.FieldNames = parts
	}

	return &fe
}

type IndexDef struct {
	FieldNames []string
	ColNames   []string
	Expr       string
}

func (e IndexDef) ToDefs() []string {
	if e.Expr != "" {
		return []string{e.Expr}
	}
	return e.FieldNames
}

func (e IndexDef) TableExpr(t *Table) *Ex {
	if len(e.Expr) != 0 {
		return t.Expr(e.Expr)
	}
	if len(e.ColNames) != 0 {
		ex := Expr("")
		ex.WriteGroup(func(ex *Ex) {
			ex.WriteExpr(t.MustCols(e.ColNames...))
		})
		return ex
	}
	ex := Expr("")
	ex.WriteGroup(func(ex *Ex) {
		ex.WriteExpr(t.MustFields(e.FieldNames...))
	})
	return ex
}

type Key struct {
	Table *Table

	Name     string
	IsUnique bool
	Method   string
	Def      IndexDef
}

func (key Key) On(table *Table) *Key {
	key.Table = table
	return &key
}

func (key Key) Using(method string) *Key {
	key.Method = method
	return &key
}

func (key *Key) T() *Table {
	return key.Table
}

func (key *Key) IsPrimary() bool {
	return key.IsUnique && key.Name == "primary" || strings.HasSuffix(key.Name, "pkey")
}

type Keys struct {
	l []*Key
}

func (keys *Keys) Clone() *Keys {
	k := &Keys{}
	keys.Range(func(key *Key, idx int) {
		k.Add(key)
	})
	return k
}

func (keys *Keys) Len() int {
	if keys == nil {
		return 0
	}
	return len(keys.l)
}

func (keys *Keys) IsEmpty() bool {
	return keys.Len() == 0
}

func (keys *Keys) Key(keyName string) (key *Key) {
	keyName = strings.ToLower(keyName)
	for i := range keys.l {
		k := keys.l[i]
		if keyName == k.Name {
			return k
		}
	}
	return nil
}

func (keys *Keys) Add(nextKeys ...*Key) {
	for i := range nextKeys {
		key := nextKeys[i]
		if key == nil {
			continue
		}
		keys.l = append(keys.l, key)
	}
}

func (keys *Keys) Range(cb func(key *Key, idx int)) {
	for i := range keys.l {
		cb(keys.l[i], i)
	}
}
