package sqlx

import (
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

type Task func(db DBExecutor) error

func (task Task) Run(db DBExecutor) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %s; calltrace:%s", fmt.Sprint(e), string(debug.Stack()))
		}
	}()
	return task(db)
}

func NewTasks(db DBExecutor) *Tasks {
	return &Tasks{
		db: db,
	}
}

type Tasks struct {
	db    DBExecutor
	tasks []Task
}

func (tasks Tasks) With(task ...Task) *Tasks {
	tasks.tasks = append(tasks.tasks, task...)
	return &tasks
}

func (tasks *Tasks) Do() (err error) {
	if len(tasks.tasks) == 0 {
		return nil
	}

	db := tasks.db

	if maybeTx, ok := db.(MaybeTxExecutor); ok {
		inTxScope := false

		if !maybeTx.IsTx() {
			db, err = maybeTx.Begin()
			if err != nil {
				return err
			}
			maybeTx = db.(MaybeTxExecutor)
			inTxScope = true
		}

		for _, task := range tasks.tasks {
			if runErr := task.Run(db); runErr != nil {
				if inTxScope {
					// err will bubble up，just handle and rollback in outermost layer
					logrus.Errorf("SQL FAILED: %s", runErr)
					if rollBackErr := maybeTx.Rollback(); rollBackErr != nil {
						logrus.Warnf("ROLLBACK FAILED: %s", rollBackErr)
						err = rollBackErr
						return
					}
				}
				return runErr
			}
		}

		if inTxScope {
			if commitErr := maybeTx.Commit(); commitErr != nil {
				logrus.Warnf("TRANSACTION COMMIT FAILED: %s", commitErr.Error())
				return commitErr
			}
		}

	}
	return nil
}
