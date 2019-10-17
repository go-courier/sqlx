package sqlx_test

import (
	"context"
	"database/sql/driver"
	"os"
	"testing"

	"github.com/go-courier/metax"
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
	logrus.AddHook(&MetaHook{})
}

type MetaHook struct {
}

func (hook *MetaHook) Fire(entry *logrus.Entry) error {
	ctx := entry.Context
	if ctx == nil {
		ctx = context.Background()
	}
	meta := metax.MetaFromContext(ctx)
	entry.Data["meta"] = meta.String()
	return nil
}

func (hook *MetaHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

type TableOperateTime struct {
	CreatedAt datatypes.MySQLDatetime `db:"f_created_at,default=CURRENT_TIMESTAMP,onupdate=CURRENT_TIMESTAMP"`
	UpdatedAt int64                   `db:"f_updated_at,default='0'"`
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
	ID       uint64 `db:"f_id,autoincrement"`
	Name     string `db:"f_name,size=255,default=''"`
	Nickname string `db:"f_nickname,size=255,default=''"`
	Username string `db:"f_username,default=''"`
	Gender   Gender `db:"f_gender,default='0'"`

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
	ID       uint64 `db:"f_id,autoincrement"`
	Nickname string `db:"f_nickname,size=255,default=''"`
	Gender   Gender `db:"f_gender,default='0'"`
	Name     string `db:"f_name,deprecated=f_real_name"`
	RealName string `db:"f_real_name,size=255,default=''"`
	Age      int32  `db:"f_age,default='0'"`
	Username string `db:"f_username,deprecated"`
}

func (user *User2) TableName() string {
	return "t_user"
}

func (user *User2) PrimaryKey() []string {
	return []string{"ID"}
}

func (user *User2) Indexes() builder.Indexes {
	return builder.Indexes{
		"I_nickname": {"Nickname"},
	}
}

func (user *User2) UniqueIndexes() builder.Indexes {
	return builder.Indexes{
		"I_name": {"RealName"},
	}
}

func TestMigrate(t *testing.T) {
	os.Setenv("PROJECT_FEATURE", "test")
	defer func() {
		os.Remove("PROJECT_FEATURE")
	}()

	dbTest := sqlx.NewFeatureDatabase("test_for_migrate")

	for _, connector := range []driver.Connector{
		mysqlConnector,
		//postgresConnector,
	} {
		for _, schema := range []string{"import", "public", "backup"} {
			t.Run("create table", func(t *testing.T) {
				dbTest.Register(&User{})
				db := dbTest.OpenDB(connector).WithSchema(schema)
				err := migration.Migrate(db, nil)
				require.NoError(t, err)
			})

			t.Run("no migrate", func(t *testing.T) {
				dbTest.Register(&User{})
				db := dbTest.OpenDB(connector).WithSchema(schema)
				err := migration.Migrate(db, nil)
				require.NoError(t, err)

				t.Run("migrate to user2", func(t *testing.T) {
					dbTest.Register(&User2{})
					db := dbTest.OpenDB(connector).WithSchema(schema)
					err := migration.Migrate(db, nil)
					require.NoError(t, err)
				})
			})

			t.Run("migrate to user", func(t *testing.T) {
				db := dbTest.OpenDB(connector).WithSchema(schema)
				err := migration.Migrate(db, nil)
				require.NoError(t, err)
			})

			dbTest.Tables.Range(func(table *builder.Table, idx int) {
				db := dbTest.OpenDB(connector).WithSchema(schema)
				_, err := db.ExecExpr(db.Dialect().DropTable(table))
				require.NoError(t, err)
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
		d := dbTest.OpenDB(connector)

		db := d.WithContext(metax.ContextWithMeta(d.Context(), metax.ParseMeta("_id=11111")))

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

		db.(*sqlx.DB).Tables.Range(func(t *builder.Table, idx int) {
			_, err := db.ExecExpr(db.Dialect().DropTable(t))
			tt.NoError(err)
		})
	}
}

type UserSet map[string]*User

func (UserSet) New() interface{} {
	return &User{}
}

func (u UserSet) Next(v interface{}) error {
	user := v.(*User)
	u[user.Name] = user
	return nil
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

		db.Tables.Range(func(t *builder.Table, idx int) {
			db.ExecExpr(db.Dialect().DropTable(t))
		})

		err := migration.Migrate(db, nil)
		tt.NoError(err)

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
			userSet := UserSet{}
			err := db.QueryExprAndScan(
				builder.Select(nil).From(table, builder.Where(table.F("Gender").Eq(GenderMale))),
				userSet,
			)
			tt.NoError(err)
			tt.Len(userSet, 10)
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
