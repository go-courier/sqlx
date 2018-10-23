package database

import (
	"database/sql/driver"
	"github.com/go-courier/sqlx/datatypes"
)

// @def primary ID
// @def index I_nickname Nickname
// @def index I_username Username
// @def unique_index I_name Name
// @def spatial_index I_geom Geom
type User struct {
	ID uint64 `db:"F_id" json:"-" sql:"bigint unsigned NOT NULL DEFAULT '0'"`
	// 姓名
	Name      string                   `db:"F_name" json:"name" sql:"varchar(255) NOT NULL DEFAULT ''"`
	Username  string                   `db:"F_username" json:"username" sql:"varchar(255) NOT NULL DEFAULT ''"`
	Nickname  string                   `db:"F_nickname" json:"nickname" sql:"varchar(255) NOT NULL DEFAULT ''"`
	Gender    Gender                   `db:"F_gender" json:"gender" sql:"int NOT NULL DEFAULT '0'"`
	Birthday  datatypes.MySQLDatetime  `db:"F_birthday" json:"birthday" sql:"datetime NOT NULL"`
	Boolean   bool                     `db:"F_boolean" json:"boolean" sql:"tinyint(1) NOT NULL DEFAULT '0'"`
	Geom      GeomString               `db:"F_geom" json:"geom" sql:"geometry NOT NULL"`
	CreatedAt datatypes.MySQLTimestamp `db:"F_created_at" json:"createdAt" sql:"bigint NOT NULL DEFAULT '0'"`
	UpdatedAt datatypes.MySQLTimestamp `db:"F_updated_at" json:"updatedAt" sql:"bigint NOT NULL DEFAULT '0'"`
	Enabled   datatypes.Bool           `db:"F_enabled" json:"enabled" sql:"int NOT NULL DEFAULT '0'"`
}

type User2 struct {
	Name string `db:"F_name" json:"name" sql:"varchar(255) NOT NULL DEFAULT ''"`
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

func (GeomString) ValueEx() string {
	return "ST_GeomFromText(?)"
}

func (GeomString) DataType() string {
	return "geometry"
}
