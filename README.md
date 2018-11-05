## Sqlx

[![GoDoc Widget](https://godoc.org/github.com/go-courier/sqlx/v2?status.svg)](https://godoc.org/github.com/go-courier/sqlx/v2)
[![Build Status](https://travis-ci.org/go-courier/sqlx.svg?branch=master)](https://travis-ci.org/go-courier/sqlx)
[![codecov](https://codecov.io/gh/go-courier/sqlx/branch/master/graph/badge.svg)](https://codecov.io/gh/go-courier/sqlx)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-courier/sqlx/v2)](https://goreportcard.com/report/github.com/go-courier/sqlx/v2)


Sql helpers just for mysql(5.7+)/postgres(10+) and mysql/postgres-compatibility db.



```go
// @def primary ID
// @def index I_nickname/BTREE Nickname
// @def index I_username Username
// @def index I_geom/SPATIAL Geom
// @def unique_index I_name Name
type User struct {
	ID uint64 `db:"F_id,autoincrement"`
	// 姓名
	Name      string                   `db:"F_name,default=''"`
	Username  string                   `db:"F_username,default=''"`
	Nickname  string                   `db:"F_nickname,default=''"`
	Gender    Gender                   `db:"F_gender,default='0'"`
	Boolean   bool                     `db:"F_boolean,default=false"`
	Geom      GeomString               `db:"F_geom"`
	CreatedAt datatypes.MySQLTimestamp `db:"F_created_at,default='0'"`
	UpdatedAt datatypes.MySQLTimestamp `db:"F_updated_at,default='0'"`
	Enabled   datatypes.Bool           `db:"F_enabled,default='0'"`
}

type GeomString struct {
	V string
}

func (g GeomString) Value() (driver.Value, error) {
	return g.V, nil
}

func (g *GeomString) Scan(src interface{}) error {
	return nil
}

func (GeomString) DataType(driverName string) string {
	if driverName == "mysql" {
		return "geometry"
	}
	return "geometry(Point)"
}

func (GeomString) ValueEx() string {
	return "ST_GeomFromText(?)"
}
```
