package mysql

import (
	"fmt"
	"github.com/go-courier/sqlx"
	"github.com/go-courier/sqlx/builder"
	"github.com/go-courier/sqlx/enummeta"
	"github.com/sirupsen/logrus"
)

type Migration struct {
	DropColumn bool
	DryRun     bool
}

func (opts Migration) Migrate(database *sqlx.Database, db *sqlx.DB) error {
	database.Register(&enummeta.SqlMetaEnum{})

	currentDatabase := DBFromInformationSchema(db, database.Name, database.Tables.TableNames()...)

	if !opts.DryRun {
		logrus.Debugf("=================== migrating database `%s` ====================", database.Name)
		defer logrus.Debugf("=================== migrated database `%s` ====================", database.Name)

		if currentDatabase == nil {
			currentDatabase = &sqlx.Database{
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

			stmt := currentTable.Diff(table, builder.DiffOptions{
				DropColumn: opts.DropColumn,
			})
			if stmt != nil {
				if _, err := db.ExecExpr(stmt); err != nil {
					return err
				}
				continue
			}
		}

		if err := enummeta.SyncEnum(database, db); err != nil {
			return err
		}

		return nil
	}

	if currentDatabase == nil {
		currentDatabase = &sqlx.Database{
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

		stmt := currentTable.Diff(table, builder.DiffOptions{
			DropColumn: opts.DropColumn,
		})

		if stmt != nil {
			fmt.Println(stmt.Expr().Query)
			continue
		}
	}

	fmt.Printf("=================== need to migrate database `%s` ====================\n", database.Name)

	if err := enummeta.SyncEnum(database, db); err != nil {
		return err
	}

	return nil
}
