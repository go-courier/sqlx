package builder

import (
	"container/list"
	"strings"
)

func PrimaryKey(columns *Columns) *Key {
	return UniqueIndex("PRIMARY", columns)
}

func Index(name string, columns *Columns) *Key {
	return &Key{
		Name:    name,
		Columns: columns,
	}
}

func UniqueIndex(name string, columns *Columns) *Key {
	return &Key{
		Name:     name,
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
	return key.T()
}

func (key *Key) IsPrimary() bool {
	return key.IsUnique && (strings.ToLower(key.Name) == "primary" || strings.HasSuffix(strings.ToLower(key.Name), "pkey"))
}

type Keys struct {
	m map[string]*list.Element
	l *list.List
}

func (keys *Keys) Clone() *Keys {
	k := &Keys{}
	keys.Range(func(key *Key, idx int) {
		k.Add(key)
	})
	return k
}

func (keys *Keys) Len() int {
	if keys.l == nil {
		return 0
	}
	return keys.l.Len()
}

func (keys *Keys) IsEmpty() bool {
	return keys.l == nil || keys.l.Len() == 0
}

func (keys *Keys) Key(keyName string) (key *Key) {
	if keys.m != nil {
		if c, ok := keys.m[strings.ToLower(keyName)]; ok {
			return c.Value.(*Key)
		}
	}
	return nil
}

func (keys *Keys) Add(nextKeys ...*Key) {
	if keys.m == nil {
		keys.m = map[string]*list.Element{}
		keys.l = list.New()
	}
	for _, key := range nextKeys {
		if key == nil {
			continue
		}
		key.Name = strings.ToLower(key.Name)
		keys.m[key.Name] = keys.l.PushBack(key)
	}
}

func (keys *Keys) Remove(name string) {
	if keys.m != nil {
		if e, exists := keys.m[name]; exists {
			keys.l.Remove(e)
			delete(keys.m, name)
		}
	}
}

func (keys *Keys) Range(cb func(key *Key, idx int)) {
	if keys.l != nil {
		i := 0
		for e := keys.l.Front(); e != nil; e = e.Next() {
			cb(e.Value.(*Key), i)
			i++
		}
	}
}
