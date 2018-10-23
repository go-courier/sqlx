package builder

import (
	"container/list"
	"fmt"
)

func AddKey(key *Key) *Expression {
	return MustJoinExpr(" ", Expr("ADD"), key.Def())
}

func DropKey(key *Key) *Expression {
	if key.Type == PRIMARY {
		return Expr(fmt.Sprintf("DROP %s KEY", key.Type))
	}
	return Expr(fmt.Sprintf("DROP INDEX %s", key.String()))
}

var _ TableDef = (*Column)(nil)

func PrimaryKey() *Key {
	return &Key{
		Name: string(PRIMARY),
		Type: PRIMARY,
	}
}

func Index(name string) *Key {
	return &Key{
		Name: name,
		Type: INDEX,
	}
}

func UniqueIndex(name string) *Key {
	return &Key{
		Name: name,
		Type: UNIQUE_INDEX,
	}
}

func SpatialIndex(name string) *Key {
	return &Key{
		Name: name,
		Type: SPATIAL_INDEX,
	}
}

type Key struct {
	Name string
	Columns
	Type keyType
}

func (key *Key) WithCols(columns ...*Column) *Key {
	key.Columns.Add(columns...)
	return key
}

func (key *Key) String() string {
	return quote(key.Name)
}

func (key *Key) IsValidDef() bool {
	return !key.Columns.IsEmpty()
}

func (key *Key) Def() *Expression {
	if key.Type == PRIMARY {
		return MustJoinExpr(" ", Expr(string(key.Type)+" KEY"), key.Columns.Group())
	}
	return MustJoinExpr(" ", Expr(string(key.Type)+" "+key.String()), key.Columns.Group())
}

type keyType string

const (
	PRIMARY      keyType = "PRIMARY"
	INDEX        keyType = "INDEX"
	UNIQUE_INDEX keyType = "UNIQUE INDEX"
	SPATIAL_INDEX keyType = "SPATIAL INDEX"
)

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

func (keys *Keys) Key(keyName string) (key *Key, exists bool) {
	if keys.m != nil {
		if c, ok := keys.m[keyName]; ok {
			return c.Value.(*Key), true
		}
	}
	return nil, false
}

func (keys *Keys) Add(nextKeys ...*Key) {
	if keys.m == nil {
		keys.m = map[string]*list.Element{}
		keys.l = list.New()
	}
	for _, key := range nextKeys {
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

func (keys Keys) Diff(targetKeys Keys) keysDiffResult {
	r := keysDiffResult{}

	ks := keys.Clone()

	targetKeys.Range(func(key *Key, idx int) {
		if currentKey, exists := ks.Key(key.Name); exists {
			if currentKey.Def().Query != key.Def().Query {
				r.keysForUpdate.Add(key)
			}
		} else {
			r.keysForAdd.Add(key)
		}
		ks.Remove(key.Name)
	})

	ks.Range(func(key *Key, idx int) {
		r.keysForDelete.Add(key)
	})

	return r
}

type keysDiffResult struct {
	keysForAdd    Keys
	keysForUpdate Keys
	keysForDelete Keys
}

func (r keysDiffResult) IsChanged() bool {
	return !r.keysForAdd.IsEmpty() || !r.keysForUpdate.IsEmpty() || !r.keysForDelete.IsEmpty()
}
