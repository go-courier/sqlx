package migration

import (
	"context"

	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/enummeta"
)

type MigrationOpts struct {
	DryRun bool
}

type contextKeyMigrationOpts int

func MigrationOptsFromContext(ctx context.Context) *MigrationOpts {
	if opts, ok := ctx.Value(contextKeyMigrationOpts(1)).(*MigrationOpts); ok {
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
	ctx := context.WithValue(db.Context(), contextKeyMigrationOpts(1), opts)

	if err := db.(sqlx.Migrator).Migrate(ctx, db); err != nil {
		return err
	}
	if err := enummeta.SyncEnum(db); err != nil {
		return err
	}
	return nil
}
