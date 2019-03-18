package database

// @def primary ID
// organization
type Org struct {
	ID   uint64 `db:"f_id,autoincrement"`
	Name string `db:"f_name,default=''"`
	// @rel User.ID
	// 关联用户
	// xxxxx
	UserID string `db:"user_id"`
}
