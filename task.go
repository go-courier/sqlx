package sqlx

import (
	"fmt"
	"runtime/debug"

	"github.com/go-courier/logr"
	"github.com/pkg/errors"
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

	log := logr.FromContext(db.Context())

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
					log.Error(errors.Wrap(err, "SQL FAILED"))
					if rollBackErr := maybeTx.Rollback(); rollBackErr != nil {
						log.Warn(errors.Wrap(rollBackErr, "ROLLBACK FAILED"))
						err = rollBackErr
						return
					}
				}
				return runErr
			}
		}

		if inTxScope {
			if commitErr := maybeTx.Commit(); commitErr != nil {
				log.Warn(errors.Wrap(commitErr, "TRANSACTION COMMIT FAILED"))
				return commitErr
			}
		}

	}
	return nil
}
