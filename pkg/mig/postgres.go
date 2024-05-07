package mig

import (
	"context"
	"database/sql"
	"github.com/alexisvisco/mig/pkg/schema"
	"github.com/alexisvisco/mig/pkg/schema/pg"
	"time"
)

// MigratePostgres migrates the database, it is launched via the generated main file.
func MigratePostgres(options *MainOptions) (bool, error) {
	var conn *sql.DB
	if options.Connection != nil {
		conn = options.Connection
	} else {
		db, err := GetConnection(options.DSN, options.Verbose)
		if err != nil {
			return false, err
		}
		conn = db
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(options.Timeout))
	defer cancel()

	migrator := schema.NewMigrator(ctx, conn, options.Tracker, pg.NewPostgres, &schema.MigratorOption{
		DryRun:             options.DryRun,
		ContinueOnError:    options.ContinueOnError,
		SchemaVersionTable: options.SchemaVersionTable,
	})

	return migrator.Apply(options.MigrationDirection, options.Version, options.Steps, options.Migrations), nil
}
