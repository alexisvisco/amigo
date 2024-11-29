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

	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
)

var (
	ErrConnectionNil   = errors.New("connection is nil")
	ErrMigrationFailed = errors.New("migration failed")
)

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

	ctx, cancel := context.WithDeadline(originCtx, time.Now().Add(a.Config.Migration.Timeout))
	defer cancel()

	a.SetupSlog(params.LogOutput, params.Logger)

	migrator, err := a.GetMigrationApplier(ctx, params.DB)
	if err != nil {
		return err
	}

	ok := migrator.Apply(
		params.Direction,
		utils.NilOrValue(a.Config.Migration.Version),
		utils.NilOrValue(a.Config.Migration.Steps),
		params.Migrations,
	)

	if !ok {
		return ErrMigrationFailed
	}

	if a.Config.Migration.DumpSchemaAfter {
		file, err := utils.CreateOrOpenFile(a.Config.SchemaOutPath)
		if err != nil {
			return fmt.Errorf("unable to open/create file: %w", err)
		}

		defer file.Close()

		err = a.DumpSchema(file, false)
		if err != nil {
			return fmt.Errorf("unable to dump schema after migrating: %w", err)
		}

		logger.Info(events.FileModifiedEvent{FileName: path.Join(a.Config.SchemaOutPath)})
	}

	return nil
}

func (a Amigo) validateRunMigration(conn *sql.DB, direction *types.MigrationDirection) error {
	if a.Config.SchemaVersionTable == "" {
		a.Config.SchemaVersionTable = amigoconfig.DefaultSchemaVersionTable
	}

	if direction == nil || *direction == "" {
		*direction = types.MigrationDirectionUp
	}

	if a.Config.Migration.Timeout == 0 {
		a.Config.Migration.Timeout = amigoconfig.DefaultTimeout
	}

	if conn == nil {
		return ErrConnectionNil
	}

	return nil
}
