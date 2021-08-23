package migration

import (
	"context"
	"io"

	contextx "github.com/go-courier/x/context"

	"github.com/go-courier/sqlx/v2"
	"github.com/go-courier/sqlx/v2/enummeta"
)

type contextKeyMigrationOutput struct{}

func MigrationOutputFromContext(ctx context.Context) io.Writer {
	if opts, ok := ctx.Value(contextKeyMigrationOutput{}).(io.Writer); ok {
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
	ctx := contextx.WithValue(db.Context(), contextKeyMigrationOutput{}, output)

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
