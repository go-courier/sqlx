package sqlx_test

import (
	"database/sql/driver"
	"os"
	"testing"

	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/sqlx/v2/datatypes"
	"github.com/go-courier/sqlx/v2/migration"
	"github.com/go-courier/sqlx/v2/mysqlconnector"
	"github.com/go-courier/sqlx/v2/postgresqlconnector"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	mysqlConnector = &mysqlconnector.MysqlConnector{
		Host:  "root@tcp(0.0.0.0:3306)",
		Extra: "charset=utf8mb4&parseTime=true&interpolateParams=true&autocommit=true&loc=Local",
	}

	postgresConnector = &postgresqlconnector.PostgreSQLConnector{
		Host:       "postgres://postgres@0.0.0.0:5432",
		Extra:      "sslmode=disable",
		Extensions: []string{"postgis"},
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
	defer func() {
		os.Remove("PROJECT_FEATURE")
	}()
	dbTest := sqlx.NewFeatureDatabase("test_for_migrate")

	for _, connector := range []driver.Connector{
		mysqlConnector,
		postgresConnector,
	} {
		for _, schema := range []string{"import", "public", "backup"} {
			db := dbTest.OpenDB(connector).WithSchema(schema)
			{
				dbTest.Register(&User{})
				err := migration.Migrate(db, nil)
				tt.NoError(err)
			}
			{
				dbTest.Register(&User{})
				err := migration.Migrate(db, nil)
				tt.NoError(err)
			}
			{
				dbTest.Register(&User2{})
				err := migration.Migrate(db, nil)
				tt.NoError(err)
			}
			{
				dbTest.Register(&User{})
				err := migration.Migrate(db, nil)
				tt.NoError(err)
			}

			dbTest.Tables.Range(func(t *builder.Table, idx int) {
				_, err := db.ExecExpr(db.Dialect().DropTable(t))
				tt.NoError(err)
			})
		}
	}
}

func TestCRUD(t *testing.T) {
	tt := require.New(t)

	dbTest := sqlx.NewDatabase("test")

	for _, connector := range []driver.Connector{
		mysqlConnector,
		postgresConnector,
	} {
		db := dbTest.OpenDB(connector)

		userTable := dbTest.Register(&User{})
		err := migration.Migrate(db, nil)
		tt.NoError(err)

		{
			user := User{
				Name:   uuid.New().String(),
				Gender: GenderMale,
			}

			_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
			tt.NoError(err)

			{
				user.Gender = GenderFemale
				_, err := db.ExecExpr(
					builder.Update(dbTest.T(&user)).
						Set(sqlx.AsAssignments(db, &user)...).
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
				_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
				t.Log(err)
				tt.True(sqlx.DBErr(err).IsConflict())
			}
		}

		db.Tables.Range(func(t *builder.Table, idx int) {
			_, err := db.ExecExpr(db.Dialect().DropTable(t))
			tt.NoError(err)
		})
	}

}

func TestSelect(t *testing.T) {
	tt := require.New(t)

	dbTest := sqlx.NewDatabase("test_for_s")

	for _, connector := range []driver.Connector{
		mysqlConnector,
		postgresConnector,
	} {
		db := dbTest.OpenDB(connector)

		table := dbTest.Register(&User{})
		err := migration.Migrate(db, nil)
		tt.Nil(err)

		for i := 0; i < 10; i++ {
			user := User{
				Name:   uuid.New().String(),
				Gender: GenderMale,
			}
			_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
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

		db.Tables.Range(func(t *builder.Table, idx int) {
			_, err := db.ExecExpr(db.Dialect().DropTable(t))
			tt.NoError(err)
		})
	}
}
