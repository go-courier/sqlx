package database

import (
	"database/sql/driver"
	"testing"

	"github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/sqlx/v2/migration"
	"github.com/go-courier/sqlx/v2/mysqlconnector"
	"github.com/go-courier/sqlx/v2/postgresqlconnector"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
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

func TestUser(t *testing.T) {
	t.Run("CRUD", func(t *testing.T) {
		for _, connector := range []driver.Connector{
			mysqlConnector,
			postgresConnector,
		} {
			db := DBTest.OpenDB(connector)

			db.D().Tables.Range(func(table *builder.Table, idx int) {
				_, err := db.ExecExpr(db.Dialect().DropTable(table))
				gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
			})

			err := migration.Migrate(db, nil)
			gomega.NewWithT(t).Expect(err).To(gomega.BeNil())

			t.Run("Create flow", func(t *testing.T) {
				user := User{}
				user.Name = uuid.New().String()
				user.Geom = GeomString{
					V: "Point(0 0)",
				}

				errForCreate := user.Create(db)
				gomega.NewWithT(t).Expect(errForCreate).To(gomega.BeNil())
				gomega.NewWithT(t).Expect(user.ID, uint64(0))

				user.Gender = GenderMale
				{
					err := user.CreateOnDuplicateWithUpdateFields(db, []string{"Gender"})
					gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
				}
				{
					userForFetch := User{
						Name: user.Name,
					}
					err := userForFetch.FetchByName(db)

					gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
					gomega.NewWithT(t).Expect(userForFetch.Gender).To(gomega.Equal(user.Gender))
				}
			})
			t.Run("delete flow", func(t *testing.T) {
				user := User{}
				user.Name = uuid.New().String()
				user.Geom = GeomString{
					V: "Point(0 0)",
				}

				errForCreate := user.Create(db)
				gomega.NewWithT(t).Expect(errForCreate).To(gomega.BeNil())

				{
					userForDelete := User{
						Name: user.Name,
					}
					err := userForDelete.SoftDeleteByName(db)
					gomega.NewWithT(t).Expect(err).To(gomega.BeNil())

					userForSelect := &User{
						Name: user.Name,
					}
					errForSelect := userForSelect.FetchByName(db)
					gomega.NewWithT(t).Expect(errForSelect).NotTo(gomega.BeNil())
				}
			})
			db.D().Tables.Range(func(table *builder.Table, idx int) {
				_, err := db.ExecExpr(db.Dialect().DropTable(table))
				gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
			})
		}
	})
	t.Run("List", func(t *testing.T) {
		for _, connector := range []driver.Connector{
			mysqlConnector,
			postgresConnector,
		} {
			db := DBTest.OpenDB(connector)

			db.D().Tables.Range(func(table *builder.Table, idx int) {
				_, err := db.ExecExpr(db.Dialect().DropTable(table))
				gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
			})

			err := migration.Migrate(db, nil)
			gomega.NewWithT(t).Expect(err).To(gomega.BeNil())

			createUser := func() {
				user := User{}
				user.Name = uuid.New().String()
				user.Geom = GeomString{
					V: "Point(0 0)",
				}

				err := user.Create(db)
				gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
			}

			for i := 0; i < 10; i++ {
				createUser()
			}

			list, err := (&User{}).List(db, nil)
			gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
			gomega.NewWithT(t).Expect(list).To(gomega.HaveLen(10))

			count, err := (&User{}).Count(db, nil)
			gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
			gomega.NewWithT(t).Expect(count).To(gomega.Equal(10))

			names := make([]string, 0)
			for _, user := range list {
				names = append(names, user.Name)
			}

			{
				list, err := (&User{}).BatchFetchByNameList(db, names)
				gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
				gomega.NewWithT(t).Expect(list).To(gomega.HaveLen(10))
			}

			db.D().Tables.Range(func(table *builder.Table, idx int) {
				_, err := db.ExecExpr(db.Dialect().DropTable(table))
				gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
			})
		}
	})
}
