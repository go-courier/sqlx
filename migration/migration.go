package migration

import (
	"github.com/go-courier/sqlx"
	"github.com/go-courier/sqlx/enummeta"
)

type Migrator interface {
	Migrate(db *sqlx.DB, database *sqlx.Database, opt *MigrationOpts) error
}

type MigrationOpts struct {
	SkipDropColumn bool
	DryRun         bool
}

func MustMigrate(db *sqlx.DB, database *sqlx.Database, opts *MigrationOpts) {
	if err := Migrate(db, database, opts); err != nil {
		panic(err)
	}
}

func Migrate(db *sqlx.DB, database *sqlx.Database, opts *MigrationOpts) error {
	database.Register(&enummeta.SqlMetaEnum{})

	if migrator, ok := db.Dialect.(Migrator); ok {
		if err := migrator.Migrate(db, database, opts); err != nil {
			return nil
		}
		if err := enummeta.SyncEnum(db, database); err != nil {
			return err
		}

	}
	return nil
}
