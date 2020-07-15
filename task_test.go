package sqlx_test

import (
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/sqlx/v2/migration"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
)

func TestWithTasks(t *testing.T) {
	dbTest := sqlx.NewDatabase("test_for_user")

	t.Run("", func(t *testing.T) {
		for _, connector := range []driver.Connector{
			mysqlConnector,
			postgresConnector,
		} {
			db := dbTest.OpenDB(connector)
			driverName := db.Dialect().DriverName()

			dbTest.Register(&User{})
			err := migration.Migrate(db, nil)
			gomega.NewWithT(t).Expect(err).To(gomega.BeNil())

			t.Run("rollback on task err", func(t *testing.T) {
				taskList := sqlx.NewTasks(db)

				taskList = taskList.With(func(db sqlx.DBExecutor) error {
					user := User{
						Name:   uuid.New().String(),
						Gender: GenderMale,
					}
					_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
					return err
				})

				taskList = taskList.With(func(db sqlx.DBExecutor) error {
					subTaskList := sqlx.NewTasks(db)

					subTaskList = subTaskList.With(func(db sqlx.DBExecutor) error {
						user := User{
							Name:   uuid.New().String(),
							Gender: GenderMale,
						}

						_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
						return err
					})

					subTaskList = subTaskList.With(func(db sqlx.DBExecutor) error {
						return fmt.Errorf("rollback")
					})

					return subTaskList.Do()
				})

				err := taskList.Do()
				gomega.NewWithT(t).Expect(err).NotTo(gomega.BeNil())
			})
			if driverName == "mysql" {
				t.Run("skip rollback", func(t *testing.T) {
					taskList := sqlx.NewTasks(db)

					user := User{
						Name:   uuid.New().String(),
						Gender: GenderMale,
					}

					taskList = taskList.With(func(db sqlx.DBExecutor) error {
						_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
						return err
					})

					taskList = taskList.With(func(db sqlx.DBExecutor) error {
						_, _ = db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
						return nil
					})

					taskList = taskList.With(func(db sqlx.DBExecutor) error {
						_, err := db.ExecExpr(sqlx.InsertToDB(db, &User{
							Name:   uuid.New().String(),
							Gender: GenderMale,
						}, nil))
						return err
					})

					err := taskList.Do()
					gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
				})
			} else {
				t.Run("skip rollback in postgres", func(t *testing.T) {
					taskList := sqlx.NewTasks(db)

					user := User{
						Name:   uuid.New().String(),
						Gender: GenderMale,
					}

					taskList = taskList.With(func(db sqlx.DBExecutor) error {
						_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
						return err
					})

					taskList = taskList.With(func(db sqlx.DBExecutor) error {
						_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil, builder.OnConflict(builder.Cols("f_name")).DoNothing()))
						return err
					})

					taskList = taskList.With(func(db sqlx.DBExecutor) error {
						_, err := db.ExecExpr(sqlx.InsertToDB(db, &User{
							Name:   uuid.New().String(),
							Gender: GenderMale,
						}, nil))
						return err
					})

					err := taskList.Do()
					gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
				})
			}

			t.Run("transaction chain", func(t *testing.T) {
				taskList := sqlx.NewTasks(db)

				taskList = taskList.With(func(db sqlx.DBExecutor) error {
					user := User{
						Name:   uuid.New().String(),
						Gender: GenderMale,
					}

					_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))

					return err
				})

				taskList = taskList.With(func(db sqlx.DBExecutor) error {
					subTaskList := sqlx.NewTasks(db)

					subTaskList = subTaskList.With(func(db sqlx.DBExecutor) error {
						user := User{
							Name:   uuid.New().String(),
							Gender: GenderMale,
						}
						_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
						return err
					})

					subTaskList = subTaskList.With(func(db sqlx.DBExecutor) error {
						user := User{
							Name:   uuid.New().String(),
							Gender: GenderMale,
						}
						_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
						return err
					})

					return subTaskList.Do()
				})

				err := taskList.Do()
				gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
			})
			db.Tables.Range(func(table *builder.Table, idx int) {
				_, err := db.ExecExpr(db.Dialect().DropTable(table))
				gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
			})
		}
	})
}
