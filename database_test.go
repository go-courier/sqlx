package sqlx

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/go-courier/sqlx/builder"
	"github.com/go-courier/sqlx/datatypes"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var db *DB

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	db = MustOpen("logger:mysql", "root@tcp(0.0.0.0:3306)/?charset=utf8&parseTime=true&interpolateParams=true&autocommit=true&loc=Local")
}

type TableOperateTime struct {
	CreatedAt datatypes.MySQLDatetime `db:"F_created_at" sql:"timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)" `
	UpdatedAt int64                   `db:"F_updated_at" sql:"bigint(64) NOT NULL DEFAULT '0'"`
}

func (t *TableOperateTime) BeforeUpdate() {
	time.Now()
	t.UpdatedAt = time.Now().UnixNano()
}

func (t *TableOperateTime) BeforeInsert() {
	t.CreatedAt = datatypes.MySQLDatetime(time.Now())
	t.UpdatedAt = t.CreatedAt.Unix()
}

type Gender int

const (
	GenderMale Gender = iota + 1
	GenderFemale
)

func (Gender) EnumType() string {
	return "Gender"
}

func (Gender) Enums() map[int][]string {
	return map[int][]string{
		int(GenderMale):   {"male", "男"},
		int(GenderFemale): {"female", "女"},
	}
}

func (g Gender) String() string {
	switch g {
	case GenderMale:
		return "male"
	case GenderFemale:
		return "female"
	}
	return ""
}

// @def primary ID
// @def index I_nickname Nickname Name
// @def unique_index I_name Name
type User struct {
	ID       uint64 `db:"F_id" sql:"bigint(64) unsigned NOT NULL AUTO_INCREMENT"`
	Name     string `db:"F_name" sql:"varchar(255) binary NOT NULL DEFAULT ''"`
	Username string `db:"F_username" sql:"varchar(255)"`
	Nickname string `db:"F_nickname" sql:"varchar(255) CHARACTER SET latin1 binary NOT NULL DEFAULT ''"`
	Gender   Gender `db:"F_gender" sql:"int(32) NOT NULL DEFAULT '0'"`

	TableOperateTime
}

func (user *User) Comments() map[string]string {
	return map[string]string{
		"Name": "姓名",
	}
}

func (user *User) AfterInsert(result sql.Result) error {
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = uint64(lastInsertID)
	return nil
}

func (user *User) TableName() string {
	return "t_user"
}

func (user *User) PrimaryKey() builder.FieldNames {
	return []string{"ID"}
}

func (user *User) Indexes() builder.Indexes {
	return builder.Indexes{
		"I_nickname": {"Nickname", "Name"},
	}
}

func (user *User) UniqueIndexes() builder.Indexes {
	return builder.Indexes{
		"I_name": {"Name"},
	}
}

type User2 struct {
	User
	Age int32 `db:"F_age" sql:"int(32) NOT NULL DEFAULT '0'"`
}

func TestMigrate(t *testing.T) {
	tt := require.New(t)

	os.Setenv("PROJECT_FEATURE", "test")
	dbTest := NewFeatureDatabase("test_for_migrate")
	defer func() {
		_, err := db.ExecExpr(builder.DropDatabase(dbTest))
		tt.NoError(err)
	}()

	{
		dbTest.Register(&User{})
		err := dbTest.MigrateTo(db, MigrationOptions{})
		tt.NoError(err)
	}
	{
		dbTest.Register(&User{})
		err := dbTest.MigrateTo(db, MigrationOptions{})
		tt.NoError(err)
	}
	{
		dbTest.Register(&User2{})
		err := dbTest.MigrateTo(db, MigrationOptions{})
		tt.NoError(err)
	}

	{
		dbTest.Register(&User{})
		err := dbTest.MigrateTo(db, MigrationOptions{})
		tt.NoError(err)
	}
}

func TestCRUD(t *testing.T) {
	tt := require.New(t)

	dbTest := NewDatabase("test")
	defer func() {
		_, err := db.ExecExpr(builder.DropDatabase(dbTest))
		tt.NoError(err)
	}()

	userTable := dbTest.Register(&User{})
	err := dbTest.MigrateTo(db, MigrationOptions{})
	tt.NoError(err)

	{
		user := User{
			Name:   uuid.New().String(),
			Gender: GenderMale,
		}
		user.BeforeInsert()
		stmt := builder.Insert().Into(dbTest.T(&user)).Set(dbTest.Assignments(&user)...)
		result, err := db.ExecExpr(stmt)
		tt.NoError(err)
		user.AfterInsert(result)
		tt.NotEmpty(user.ID)

		{
			user.Gender = GenderFemale
			user.BeforeUpdate()
			_, err := db.ExecExpr(
				builder.Update(dbTest.T(&user)).
					Set(dbTest.Assignments(&user)...).
					Where(
						userTable.F("Name").Eq(user.Name),
					),
			)
			tt.Nil(err)
		}

		{
			userForSelect := User{}
			err := db.QueryExprAndScan(
				builder.Select(nil).From(
					userTable,
					builder.Where(userTable.F("Name").Eq(user.Name)),
					builder.Comment("FindUser"),
				),
				&userForSelect)

			tt.NoError(err)
			tt.Equal(userForSelect.Name, user.Name)
			tt.Equal(userForSelect.CreatedAt.Unix(), user.CreatedAt.Unix())
		}

		{
			user.BeforeInsert()
			_, err := db.ExecExpr(builder.Insert().Into(dbTest.T(&user)).Set(dbTest.Assignments(&user)...))
			t.Log(err)
			tt.True(DBErr(err).IsConflict())

			{
				_, err := db.ExecExpr(
					builder.Insert().Into(dbTest.T(&user),
						builder.OnDuplicateKeyUpdate(
							userTable.AssignmentsByFieldValues(builder.FieldValues{
								"Gender": GenderMale,
							})...,
						),
						builder.Comment("InsertUserOnDuplicate"),
					).Set(dbTest.Assignments(&user)...),
				)
				tt.Nil(err)
			}
		}
	}

}

func TestSelect(t *testing.T) {
	tt := require.New(t)

	dbTest := NewDatabase("test2")
	defer func() {
		_, err := db.ExecExpr(builder.DropDatabase(dbTest))
		tt.Nil(err)
	}()

	table := dbTest.Register(&User{})
	err := dbTest.MigrateTo(db, MigrationOptions{})
	tt.Nil(err)

	for i := 0; i < 10; i++ {
		user := User{
			Name:   uuid.New().String(),
			Gender: GenderMale,
		}
		user.BeforeInsert()
		_, err := db.ExecExpr(builder.Insert().Into(dbTest.T(&user)).Set(dbTest.Assignments(&user)...))
		tt.NoError(err)
	}

	{
		users := make([]User, 0)
		err := db.QueryExprAndScan(
			builder.Select(nil).From(table, builder.Where(table.F("Gender").Eq(GenderMale))),
			&users,
		)
		tt.NoError(err)
		tt.Len(users, 10)
	}
	{
		user := User{}
		err := db.QueryExprAndScan(
			builder.Select(nil).From(
				table,
				builder.Where(table.F("ID").Eq(11)),
			),
			&user,
		)
		tt.True(DBErr(err).IsNotFound())
	}
	{
		count := 0
		err := db.QueryExprAndScan(
			builder.Select(builder.Count()).From(table),
			&count,
		)
		tt.NoError(err)
		tt.Equal(10, count)
	}
	{
		user := &User{}
		err := db.QueryExprAndScan(
			builder.Select(builder.Count()).From(
				table,
				builder.Where(table.F("Gender").Eq(GenderMale)),
			),
			&user,
		)
		tt.Error(err)
	}
}
