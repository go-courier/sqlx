package migration

import (
	"context"
	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/enummeta"
)

type MigrationOpts struct {
	SkipDropColumn bool
	DryRun         bool
}

var contextKeyMigrationOpts = "#####MigrationOpts#####"

func MigrationOptsFromContext(ctx context.Context) *MigrationOpts {
	if opts, ok := ctx.Value(contextKeyMigrationOpts).(*MigrationOpts); ok {
		if opts != nil {
			return opts
		}
	}
	return &MigrationOpts{}
}

func MustMigrate(db sqlx.DBExecutor, opts *MigrationOpts) {
	if err := Migrate(db, opts); err != nil {
		panic(err)
	}
}

func Migrate(db sqlx.DBExecutor, opts *MigrationOpts) error {
	ctx := context.WithValue(db.Context(), contextKeyMigrationOpts, opts)

	if err := db.(sqlx.Migrator).Migrate(ctx, db); err != nil {
		return err
	}
	if err := enummeta.SyncEnum(db); err != nil {
		return err
	}
	return nil
}
