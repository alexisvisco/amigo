package amigo

import (
	"context"
	"database/sql"
	"github.com/alexisvisco/amigo/pkg/amigoctx"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils/dblog"
	"log/slog"
	"os"
	"time"
)

// RunPostgresMigrations migrates the database, it is launched via the generated main file or manually in a codebase.
func RunPostgresMigrations(options *RunMigrationOptions) (bool, error) {
	var (
		db       *sql.DB
		dblogger *dblog.Logger
		err      error
		conn     *sql.DB
	)

	if options.SchemaVersionTable == "" {
		options.SchemaVersionTable = schema.TableName(amigoctx.DefaultSchemaVersionTable)
	}

	if options.MigrationDirection == "" {
		options.MigrationDirection = types.MigrationDirectionUp
	}

	if options.Timeout == 0 {
		options.Timeout = amigoctx.DefaultTimeout
	}

	if options.Connection != nil {
		conn = options.Connection
	} else {
		db, dblogger, err = GetConnection(options.DSN)
		if err != nil {
			return false, err
		}
		conn = db
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(options.Timeout))
	defer cancel()

	migrator := schema.NewMigrator(ctx, conn, pg.NewPostgres, &schema.MigratorOption{
		DryRun:             options.DryRun,
		ContinueOnError:    options.ContinueOnError,
		SchemaVersionTable: options.SchemaVersionTable,
		DBLogger:           dblogger,
	})

	slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return migrator.Apply(options.MigrationDirection, options.Version, options.Steps, options.Migrations), nil

}
