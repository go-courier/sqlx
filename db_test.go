package sqlx_test

import (
	"fmt"
	"testing"

	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/migration"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWithTasks(t *testing.T) {
	tt := require.New(t)

	dbTest := sqlx.NewDatabase("test_for_user")
	db := dbTest.OpenDB(mysqlConnector)

	defer func() {
		_, err := db.ExecExpr(db.DropDatabase(dbTest.Name))
		tt.NoError(err)
	}()

	{
		dbTest.Register(&User{})
		err := migration.Migrate(db, nil)
		tt.NoError(err)
	}

	{
		taskList := sqlx.NewTasks(db)

		taskList = taskList.With(func(db *sqlx.DB) error {
			user := User{
				Name:   uuid.New().String(),
				Gender: GenderMale,
			}
			_, err := db.ExecExpr(dbTest.Insert(&user, nil))
			return err
		})

		taskList = taskList.With(func(db *sqlx.DB) error {
			subTaskList := sqlx.NewTasks(db)

			subTaskList = subTaskList.With(func(db *sqlx.DB) error {
				user := User{
					Name:   uuid.New().String(),
					Gender: GenderMale,
				}

				_, err := db.ExecExpr(dbTest.Insert(&user, nil))
				return err
			})

			subTaskList = subTaskList.With(func(db *sqlx.DB) error {
				return fmt.Errorf("rollback")
			})

			return subTaskList.Do()
		})

		err := taskList.Do()
		tt.NotNil(err)
	}

	taskList := sqlx.NewTasks(db)

	taskList = taskList.With(func(db *sqlx.DB) error {
		user := User{
			Name:   uuid.New().String(),
			Gender: GenderMale,
		}

		_, err := db.ExecExpr(dbTest.Insert(&user, nil))

		return err
	})

	taskList = taskList.With(func(db *sqlx.DB) error {
		subTaskList := sqlx.NewTasks(db)

		subTaskList = subTaskList.With(func(db *sqlx.DB) error {
			user := User{
				Name:   uuid.New().String(),
				Gender: GenderMale,
			}
			_, err := db.ExecExpr(dbTest.Insert(&user, nil))
			return err
		})

		subTaskList = subTaskList.With(func(db *sqlx.DB) error {
			user := User{
				Name:   uuid.New().String(),
				Gender: GenderMale,
			}
			_, err := db.ExecExpr(dbTest.Insert(&user, nil))
			return err
		})

		return subTaskList.Do()
	})

	err := taskList.Do()
	tt.NoError(err)
}
