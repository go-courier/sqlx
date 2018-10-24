package sqlx_test

import (
	"fmt"
	"github.com/go-courier/sqlx"
	"github.com/go-courier/sqlx/migration/mysql"
	"testing"

	"github.com/go-courier/sqlx/builder"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWithTasks(t *testing.T) {
	tt := require.New(t)

	dbTest := sqlx.NewDatabase("test_for_user")
	defer func() {
		_, err := db.ExecExpr(builder.DropDatabase(dbTest))
		tt.NoError(err)
	}()

	{
		dbTest.Register(&User{})
		err := (mysql.Migration{DryRun: false}).Migrate(dbTest, db)
		tt.NoError(err)
	}

	{
		taskList := sqlx.NewTasks(db)

		taskList = taskList.With(func(db *sqlx.DB) error {
			user := User{
				Name:   uuid.New().String(),
				Gender: GenderMale,
			}

			_, err := db.ExecExpr(
				builder.Insert().Into(dbTest.T(&user)).Set(dbTest.Assignments(&user)...),
			)

			return err
		})

		taskList = taskList.With(func(db *sqlx.DB) error {
			subTaskList := sqlx.NewTasks(db)

			subTaskList = subTaskList.With(func(db *sqlx.DB) error {
				user := User{
					Name:   uuid.New().String(),
					Gender: GenderMale,
				}

				_, err := db.ExecExpr(
					builder.Insert().Into(dbTest.T(&user)).Set(dbTest.Assignments(&user)...),
				)

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

		_, err := db.ExecExpr(
			builder.Insert().Into(dbTest.T(&user)).Set(dbTest.Assignments(&user)...),
		)

		return err
	})

	taskList = taskList.With(func(db *sqlx.DB) error {
		subTaskList := sqlx.NewTasks(db)

		subTaskList = subTaskList.With(func(db *sqlx.DB) error {
			user := User{
				Name:   uuid.New().String(),
				Gender: GenderMale,
			}

			_, err := db.ExecExpr(
				builder.Insert().Into(dbTest.T(&user)).Set(dbTest.Assignments(&user)...),
			)

			return err
		})

		subTaskList = subTaskList.With(func(db *sqlx.DB) error {
			user := User{
				Name:   uuid.New().String(),
				Gender: GenderMale,
			}

			_, err := db.ExecExpr(
				builder.Insert().Into(dbTest.T(&user)).Set(dbTest.Assignments(&user)...),
			)

			return err
		})

		return subTaskList.Do()
	})

	err := taskList.Do()
	tt.NoError(err)
}
