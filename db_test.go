package sqlx

import (
	"fmt"
	"testing"

	"github.com/go-courier/sqlx/builder"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWithTasks(t *testing.T) {
	tt := require.New(t)

	dbTest := NewDatabase("test_for_user")
	defer func() {
		_, err := db.ExecExpr(builder.DropDatabase(dbTest))
		tt.NoError(err)
	}()

	{
		dbTest.Register(&User{})
		err := dbTest.MigrateTo(db, MigrationOptions{DryRun: false})
		tt.NoError(err)
	}

	{
		taskList := NewTasks(db)

		taskList = taskList.With(func(db *DB) error {
			user := User{
				Name:   uuid.New().String(),
				Gender: GenderMale,
			}

			_, err := db.ExecExpr(
				builder.Insert().Into(dbTest.T(&user)).Set(dbTest.Assignments(&user)...),
			)

			return err
		})

		taskList = taskList.With(func(db *DB) error {
			subTaskList := NewTasks(db)

			subTaskList = subTaskList.With(func(db *DB) error {
				user := User{
					Name:   uuid.New().String(),
					Gender: GenderMale,
				}

				_, err := db.ExecExpr(
					builder.Insert().Into(dbTest.T(&user)).Set(dbTest.Assignments(&user)...),
				)

				return err
			})

			subTaskList = subTaskList.With(func(db *DB) error {
				return fmt.Errorf("rollback")
			})

			return subTaskList.Do()
		})

		err := taskList.Do()
		tt.NotNil(err)
	}

	taskList := NewTasks(db)

	taskList = taskList.With(func(db *DB) error {
		user := User{
			Name:   uuid.New().String(),
			Gender: GenderMale,
		}

		_, err := db.ExecExpr(
			builder.Insert().Into(dbTest.T(&user)).Set(dbTest.Assignments(&user)...),
		)

		return err
	})

	taskList = taskList.With(func(db *DB) error {
		subTaskList := NewTasks(db)

		subTaskList = subTaskList.With(func(db *DB) error {
			user := User{
				Name:   uuid.New().String(),
				Gender: GenderMale,
			}

			_, err := db.ExecExpr(
				builder.Insert().Into(dbTest.T(&user)).Set(dbTest.Assignments(&user)...),
			)

			return err
		})

		subTaskList = subTaskList.With(func(db *DB) error {
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
