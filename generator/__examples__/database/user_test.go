package database

import (
	"database/sql/driver"
	"testing"

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

func TestUserCRUD(t *testing.T) {
	tt := require.New(t)

	for _, connector := range []driver.Connector{
		mysqlConnector,
		postgresConnector,
	} {
		db := DBTest.OpenDB(connector)

		for _, t := range DBTest.Tables {
			_, err := db.ExecExpr(db.DropTable(t))
			tt.NoError(err)
		}

		err := migration.Migrate(db, nil)
		tt.NoError(err)

		user := User{}
		user.Name = uuid.New().String()
		user.Geom = GeomString{
			V: "Point(0 0)",
		}

		errForCreate := user.Create(db)
		tt.NoError(errForCreate)
		tt.Equal(uint64(0), user.ID)

		user.Gender = GenderMale
		{
			err := user.CreateOnDuplicateWithUpdateFields(db, []string{"Gender"})
			tt.NoError(err)
		}
		{
			userForFetch := User{
				Name: user.Name,
			}
			err := userForFetch.FetchByName(db)
			tt.NoError(err)

			tt.Equal(user.Gender, userForFetch.Gender)
		}
		{
			{
				userForDelete := User{
					Name: user.Name,
				}
				err := userForDelete.SoftDeleteByName(db)
				tt.NoError(err)

				userForSelect := User{
					Name: user.Name,
				}
				stmt := builder.Select(nil).From(db.T(userForSelect), builder.Where(
					userForSelect.FieldName().Eq(userForSelect.Name),
				))
				errForSelect := db.QueryExprAndScan(stmt, &userForSelect)
				tt.NoError(errForSelect)
				tt.Equal(datatypes.BOOL_FALSE, userForSelect.Enabled)

				{
					err := user.Create(db)
					tt.NoError(err)
					tt.Equal(uint64(3), user.ID)

					userForDelete := User{}
					errForSoftDelete := userForDelete.SoftDeleteByName(db)
					tt.Nil(errForSoftDelete)

					users := make([]User, 0)
					stmt := builder.Select(nil).From(db.T(userForSelect), builder.Where(
						userForSelect.FieldEnabled().Eq(datatypes.BOOL_FALSE),
					))

					errForSelect := db.QueryExprAndScan(stmt, &users)
					tt.Nil(errForSelect)
					tt.Len(users, 1)
					tt.Equal(uint64(1), users[0].ID)
				}
			}
		}

		for _, t := range DBTest.Tables {
			_, err := db.ExecExpr(db.DropTable(t))
			tt.NoError(err)
		}
	}
}

func TestUserList(t *testing.T) {
	tt := require.New(t)

	for _, connector := range []driver.Connector{
		mysqlConnector,
		postgresConnector,
	} {
		db := DBTest.OpenDB(connector)

		err := migration.Migrate(db, nil)
		tt.NoError(err)

		createUser := func() {
			user := User{}
			user.Name = uuid.New().String()
			user.Geom = GeomString{
				V: "Point(0 0)",
			}

			err := user.Create(db)
			tt.NoError(err)
		}

		for i := 0; i < 10; i++ {
			createUser()
		}

		list, err := (&User{}).List(db, nil)
		tt.NoError(err)
		tt.Len(list, 10)

		count, err := (&User{}).Count(db, nil)
		tt.NoError(err)
		tt.Equal(10, count)

		names := make([]string, 0)
		for _, user := range list {
			names = append(names, user.Name)
		}

		{
			list, err := (&User{}).BatchFetchByNameList(db, names)
			tt.NoError(err)
			tt.Len(list, 10)
		}

		for _, t := range DBTest.Tables {
			_, err := db.ExecExpr(db.DropTable(t))
			tt.NoError(err)
		}
	}
}
