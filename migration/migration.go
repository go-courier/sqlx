package migration

import (
	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/enummeta"
)

type Migrator interface {
	Migrate(db *sqlx.DB, opt *MigrationOpts) error
}

type MigrationOpts struct {
	SkipDropColumn bool
	DryRun         bool
}

func MustMigrate(db *sqlx.DB, opts *MigrationOpts) {
	if err := Migrate(db, opts); err != nil {
		panic(err)
	}
}

func Migrate(db *sqlx.DB, opts *MigrationOpts) error {
	if migrator, ok := db.Dialect.(Migrator); ok {
		if err := migrator.Migrate(db, opts); err != nil {
			return err
		}
	}

	if err := enummeta.SyncEnum(db); err != nil {
		return err
	}
	return nil
}
