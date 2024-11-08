package amigo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"
	"time"

	"github.com/alexisvisco/amigo/pkg/amigoctx"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/base"
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"github.com/alexisvisco/amigo/pkg/schema/sqlite"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/dblog"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	sqldblogger "github.com/simukti/sqldb-logger"
)

var (
	ErrConnectionNil   = errors.New("connection is nil")
	ErrMigrationFailed = errors.New("migration failed")
)

type migrationApplier interface {
	Apply(direction types.MigrationDirection, version *string, steps *int, migrations []schema.Migration) bool
}

type RunMigrationParams struct {
	DB         *sql.DB
	Direction  types.MigrationDirection
	Migrations []schema.Migration
	LogOutput  io.Writer
	Context    context.Context
	Logger     *slog.Logger
}

// RunMigrations migrates the database, it is launched via the generated main file or manually in a codebase.
func (a Amigo) RunMigrations(params RunMigrationParams) error {
	err := a.validateRunMigration(params.DB, &params.Direction)
	if err != nil {
		return err
	}

	originCtx := context.Background()
	if params.Context != nil {
		originCtx = params.Context
	}

	ctx, cancel := context.WithDeadline(originCtx, time.Now().Add(a.ctx.Migration.Timeout))
	defer cancel()

	a.SetupSlog(params.LogOutput, params.Logger)

	migrator, err := a.getMigrationApplier(ctx, params.DB)
	if err != nil {
		return err
	}

	ok := migrator.Apply(
		params.Direction,
		utils.NilOrValue(a.ctx.Migration.Version),
		utils.NilOrValue(a.ctx.Migration.Steps),
		params.Migrations,
	)

	if !ok {
		return ErrMigrationFailed
	}

	if a.ctx.Migration.DumpSchemaAfter {
		file, err := utils.CreateOrOpenFile(a.ctx.SchemaOutPath)
		if err != nil {
			return fmt.Errorf("unable to open/create file: %w", err)
		}

		defer file.Close()

		err = a.DumpSchema(file, false)
		if err != nil {
			return fmt.Errorf("unable to dump schema after migrating: %w", err)
		}

		logger.Info(events.FileModifiedEvent{FileName: path.Join(a.ctx.SchemaOutPath)})
	}

	return nil
}

func (a Amigo) validateRunMigration(conn *sql.DB, direction *types.MigrationDirection) error {
	if a.ctx.SchemaVersionTable == "" {
		a.ctx.SchemaVersionTable = amigoctx.DefaultSchemaVersionTable
	}

	if direction == nil || *direction == "" {
		*direction = types.MigrationDirectionUp
	}

	if a.ctx.Migration.Timeout == 0 {
		a.ctx.Migration.Timeout = amigoctx.DefaultTimeout
	}

	if conn == nil {
		return ErrConnectionNil
	}

	return nil
}

func (a Amigo) getMigrationApplier(
	ctx context.Context,
	conn *sql.DB,
) (migrationApplier, error) {
	recorder := dblog.NewHandler(a.ctx.ShowSQLSyntaxHighlighting)
	recorder.ToggleLogger(true)

	if a.ctx.ValidateDSN() == nil {
		conn = sqldblogger.OpenDriver(a.ctx.GetRealDSN(), conn.Driver(), recorder)
	}

	opts := &schema.MigratorOption{
		DryRun:             a.ctx.Migration.DryRun,
		ContinueOnError:    a.ctx.Migration.ContinueOnError,
		SchemaVersionTable: schema.TableName(a.ctx.SchemaVersionTable),
		DBLogger:           recorder,
		DumpSchemaFilePath: utils.NilOrValue(a.ctx.SchemaOutPath),
		UseSchemaDump:      a.ctx.Migration.UseSchemaDump,
	}

	switch a.Driver {
	case types.DriverPostgres:
		return schema.NewMigrator(ctx, conn, pg.NewPostgres, opts), nil
	case types.DriverSQLite:
		return schema.NewMigrator(ctx, conn, sqlite.NewSQLite, opts), nil
	}

	return schema.NewMigrator(ctx, conn, base.NewBase, opts), nil
}
