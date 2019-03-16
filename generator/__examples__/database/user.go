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
	ID uint64 `db:"f_id,autoincrement"`
	// 姓名
	Name      string                   `db:"f_name,default=''"`
	Username  string                   `db:"f_username,default=''"`
	Nickname  string                   `db:"f_nickname,default=''"`
	Gender    Gender                   `db:"f_gender,default='0'"`
	Boolean   bool                     `db:"f_boolean,default=false"`
	Geom      GeomString               `db:"f_geom"`
	CreatedAt datatypes.MySQLTimestamp `db:"f_created_at,default='0'"`
	UpdatedAt datatypes.MySQLTimestamp `db:"f_updated_at,default='0'"`
	Enabled   datatypes.Bool           `db:"f_enabled,default='0'"`
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
