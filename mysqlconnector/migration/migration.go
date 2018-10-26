package migration

import (
	"github.com/go-courier/sqlx"
	"github.com/go-courier/sqlx/builder"
	"github.com/go-courier/sqlx/enummeta"
	"log"
)

type Migration struct {
	Database       *sqlx.Database
	SkipDropColumn bool
	DryRun         bool
}

func (opts Migration) Migrate(db *sqlx.DB) error {
	opts.Database.Register(&enummeta.SqlMetaEnum{})

	prevDB := DBFromInformationSchema(db, opts.Database.Name, opts.Database.Tables.TableNames()...)

	if prevDB == nil {
		prevDB = &sqlx.Database{
			Name: opts.Database.Name,
		}
		if _, err := db.ExecExpr(db.CreateDatabaseIfNotExists(prevDB.Name)); err != nil {
			return err
		}
	}

	for name, table := range opts.Database.Tables {
		prevTable := prevDB.Table(name)
		if prevTable == nil {
			for _, expr := range db.CreateTableIsNotExists(table) {
				if _, err := db.ExecExpr(expr); err != nil {
					return err
				}
			}
			continue
		}

		exprList := table.Diff(prevTable, db.Dialect, opts.SkipDropColumn)

		for _, expr := range exprList {
			if !expr.IsNil() {
				if opts.DryRun {
					log.Printf(builder.ExprFrom(expr).Flatten().Query())
				} else {
					if _, err := db.ExecExpr(expr); err != nil {
						return err
					}
				}
			}
		}
	}

	if err := enummeta.SyncEnum(opts.Database, db); err != nil {
		return err
	}

	return nil
}
