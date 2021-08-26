package builder

import (
	"strings"
)

func PrimaryKey(columns *Columns) *Key {
	return UniqueIndex("PRIMARY", columns)
}

func Index(name string, columns *Columns) *Key {
	return &Key{
		Name:    strings.ToLower(name),
		Columns: columns,
	}
}

func UniqueIndex(name string, columns *Columns) *Key {
	return &Key{
		Name:     strings.ToLower(name),
		IsUnique: true,
		Columns:  columns,
	}
}

var _ TableDefinition = (*Key)(nil)

type Key struct {
	Columns *Columns
	Table   *Table

	Name     string
	IsUnique bool
	Method   string
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
