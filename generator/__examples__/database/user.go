package database

import (
	"database/sql/driver"

	"github.com/go-courier/sqlx/v2/datatypes"
)

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
