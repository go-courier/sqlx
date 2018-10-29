package sqlx_test

import (
	"database/sql"
	"github.com/go-courier/sqlx"
	"github.com/go-courier/sqlx/builder"
	"github.com/go-courier/sqlx/datatypes"
	"github.com/go-courier/sqlx/migration"
	"github.com/go-courier/sqlx/mysqlconnector"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var (
	mysqlConnector = &mysqlconnector.MysqlConnector{
		Host:  "root@tcp(0.0.0.0:3306)",
		Extra: "charset=utf8mb4&parseTime=true&interpolateParams=true&autocommit=true&loc=Local",
	}
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

type TableOperateTime struct {
	CreatedAt datatypes.MySQLDatetime `db:"F_created_at,default=CURRENT_TIMESTAMP,onupdate=CURRENT_TIMESTAMP"`
	UpdatedAt int64                   `db:"F_updated_at,default='0'"`
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

type User struct {
	ID       uint64 `db:"F_id,autoincrement"`
	Name     string `db:"F_name,size=255,default=''"`
	Nickname string `db:"F_nickname,size=255,default=''"`
	Username string `db:"F_username,default=''"`
	Gender   Gender `db:"F_gender,default='0'"`

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

func (user *User) PrimaryKey() []string {
	return []string{"ID"}
}

func (user *User) Indexes() builder.Indexes {
	return builder.Indexes{
		"I_nickname": {"Nickname"},
	}
}

func (user *User) UniqueIndexes() builder.Indexes {
	return builder.Indexes{
		"I_name": {"Name"},
	}
}

type User2 struct {
	User
	Age int32 `db:"F_age,default='0'"`
}

func TestMigrate(t *testing.T) {
	tt := require.New(t)
	os.Setenv("PROJECT_FEATURE", "test")

	dbTest := sqlx.NewFeatureDatabase("test_for_migrate")
	db := dbTest.OpenDB(mysqlConnector)

	defer func() {
		_, err := db.ExecExpr(db.DropDatabase(dbTest.Name))
		tt.NoError(err)
	}()

	{
		dbTest.Register(&User{})
		err := migration.Migrate(db, dbTest, nil)
		tt.NoError(err)
	}
	{
		dbTest.Register(&User{})
		err := migration.Migrate(db, dbTest, nil)
		tt.NoError(err)
	}
	{
		dbTest.Register(&User2{})
		err := migration.Migrate(db, dbTest, nil)
		tt.NoError(err)
	}

	{
		dbTest.Register(&User{})
		err := migration.Migrate(db, dbTest, nil)
		tt.NoError(err)
	}
}

func TestCRUD(t *testing.T) {
	tt := require.New(t)

	dbTest := sqlx.NewDatabase("test")
	db := dbTest.OpenDB(mysqlConnector)

	defer func() {
		_, err := db.ExecExpr(db.DropDatabase(dbTest.Name))
		tt.NoError(err)
	}()

	userTable := dbTest.Register(&User{})
	err := migration.Migrate(db, dbTest, nil)
	tt.NoError(err)

	{
		user := User{
			Name:   uuid.New().String(),
			Gender: GenderMale,
		}

		result, err := db.ExecExpr(dbTest.Insert(&user))
		tt.NoError(err)
		user.AfterInsert(result)
		tt.NotEmpty(user.ID)

		{
			user.Gender = GenderFemale
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
			tt.Equal(userForSelect.Gender, user.Gender)
		}

		{
			_, err := db.ExecExpr(dbTest.Insert(&user))
			t.Log(err)
			tt.True(sqlx.DBErr(err).IsConflict())
		}
	}

}

func TestSelect(t *testing.T) {
	tt := require.New(t)

	dbTest := sqlx.NewDatabase("test2")
	db := dbTest.OpenDB(mysqlConnector)

	defer func() {
		_, err := db.ExecExpr(db.DropDatabase(dbTest.Name))
		tt.Nil(err)
	}()

	table := dbTest.Register(&User{})
	err := migration.Migrate(db, dbTest, nil)
	tt.Nil(err)

	for i := 0; i < 10; i++ {
		user := User{
			Name:   uuid.New().String(),
			Gender: GenderMale,
		}
		_, err := db.ExecExpr(dbTest.Insert(&user))
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
		tt.True(sqlx.DBErr(err).IsNotFound())
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
