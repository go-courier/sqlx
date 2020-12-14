package migration

import (
	"context"
	"github.com/go-courier/sqlx/v2/enummeta"
	"io"

	"github.com/go-courier/sqlx/v2"
)

type contextKeyMigrationOutput int

func MigrationOutputFromContext(ctx context.Context) io.Writer {
	if opts, ok := ctx.Value(contextKeyMigrationOutput(1)).(io.Writer); ok {
		if opts != nil {
			return opts
		}
	}
	return nil
}

func MustMigrate(db sqlx.DBExecutor, w io.Writer) {
	if err := Migrate(db, w); err != nil {
		panic(err)
	}
}

func Migrate(db sqlx.DBExecutor, output io.Writer) error {
	ctx := context.WithValue(db.Context(), contextKeyMigrationOutput(1), output)

	if err := db.(sqlx.Migrator).Migrate(ctx, db); err != nil {
		return err
	}
	if output == nil {
		if err := enummeta.SyncEnum(db); err != nil {
			return err
		}
	}
	return nil
}
