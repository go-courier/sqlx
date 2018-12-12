package sqlx

import (
	"fmt"
	"github.com/go-courier/sqlx/builder"
	"github.com/sirupsen/logrus"
)

type MigrationOptions struct {
	DropColumn bool
	DryRun     bool
}

func (database *Database) MigrateTo(db *DB, opts MigrationOptions) error {
	database.Register(&SqlMetaEnum{})

	currentDatabase := DBFromInformationSchema(db, database.Name, database.Tables.TableNames()...)

	if !opts.DryRun {
		logrus.Debugf("=================== migrating database `%s` ====================", database.Name)
		defer logrus.Debugf("=================== migrated database `%s` ====================", database.Name)

		if currentDatabase == nil {
			currentDatabase = &Database{
				Database: builder.DB(database.Name),
			}
			if _, err := db.ExecExpr(builder.CreateDatebaseIfNotExists(currentDatabase)); err != nil {
				return err

			}
		}

		for name, table := range database.Tables {
			currentTable := currentDatabase.Table(name)
			if currentTable == nil {
				if _, err := db.ExecExpr(builder.CreateTableIsNotExists(table)); err != nil {
					return err
				}
				continue
			}

			stmts := currentTable.Diff(table, builder.DiffOptions{
				DropColumn: opts.DropColumn,
			})
			if len(stmts) > 0 {
				for i := range stmts {
					if _, err := db.ExecExpr(stmts[i]); err != nil {
						return err
					}
				}
				continue
			}
		}

		if err := SyncEnum(database, db); err != nil {
			return err
		}

		return nil
	}

	if currentDatabase == nil {
		currentDatabase = &Database{
			Database: builder.DB(database.Name),
		}

		fmt.Printf("=================== need to migrate database `%s` ====================\n", database.Name)
		fmt.Println(builder.CreateDatebase(currentDatabase).Expr().Query)
	}

	for name, table := range database.Tables {
		currentTable := currentDatabase.Table(name)
		if currentTable == nil {
			fmt.Println(builder.CreateTableIsNotExists(table).Expr().Query)
			continue
		}

		stmts := currentTable.Diff(table, builder.DiffOptions{
			DropColumn: opts.DropColumn,
		})

		if len(stmts) > 0 {
			for i := range stmts {
				fmt.Println(stmts[i].Expr().Query)
			}
			continue
		}
	}

	fmt.Printf("=================== need to migrate database `%s` ====================\n", database.Name)

	if err := SyncEnum(database, db); err != nil {
		return err
	}

	return nil
}
