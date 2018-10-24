package loggermysql

import (
	"database/sql/driver"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

type loggingTx struct {
	logger *logrus.Logger
	tx     driver.Tx
}

func (tx *loggingTx) Commit() error {
	if err := tx.tx.Commit(); err != nil {
		tx.logger.Debugf("failed to commit transaction: %s", err)
		return err
	}
	tx.logger.Debugf(color.YellowString("=========== Committed Transaction ==========="))
	return nil
}

func (tx *loggingTx) Rollback() error {
	if err := tx.tx.Rollback(); err != nil {
		tx.logger.Debugf("failed to rollback transaction: %s", err)
		return err
	}
	tx.logger.Debugf("=========== Rollback Transaction ===========")
	return nil
}
