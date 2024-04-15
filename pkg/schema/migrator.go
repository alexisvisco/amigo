package schema

import (
	"context"
	"database/sql"
	"github.com/lmittmann/tint"
	"log/slog"
	"os"
	"time"
)

// MigrationType is the context of the migrator.
type MigrationType string

const (
	// MigrationTypeUp is the up migration.
	MigrationTypeUp MigrationType = "up"

	// MigrationTypeDown is the down migration.
	MigrationTypeDown MigrationType = "down"
)

// MigratorOption is the option of the migrator.
type MigratorOption struct {
	// DryRun specifies if the migrator should perform the migrations without actually applying them.
	// Not compatible with TransactionNone.
	DryRun bool

	// ContinueOnError specifies if the migrator should continue running migrations even if an error occurs.
	// If you need to stop the migration process when an error occurs, wrap your error with NewForceStopError.
	ContinueOnError bool
}

// Migration is the interface that describes a migration at is simplest form.
type Migration interface {
	Name() string
	Date() time.Time
}

// DetailedMigration is the interface that describes a migration with up and down operations.
type DetailedMigration[T any] interface {
	Up(T)
	Down(T)
	Name() string
	Date() time.Time
}

// SimpleMigration is the interface that describes a migration with a single operation.
type SimpleMigration[T any] interface {
	Change(T)
	Name() string
	Date() time.Time
}

type SchemaFactory[T any] func(MigratorContext, DB) T

// Migrator applies the migrations.
type Migrator[T any] struct {
	db  *sql.DB
	ctx MigratorContext

	schemaFactory SchemaFactory[T]
	migrations    []func(T)
}

// NewMigrator creates a new migrator.
func NewMigrator[T any](
	ctx context.Context,
	db *sql.DB,
	schemaFactory SchemaFactory[T],
	opts *MigratorOption,
) *Migrator[T] {
	return &Migrator[T]{
		db:            db,
		schemaFactory: schemaFactory,
		ctx: MigratorContext{
			Context: ctx,
			opts:    opts,
			Logger:  slog.New(tint.NewHandler(os.Stdout, &tint.Options{})),
		},
	}
}

// run runs the migration.
func (m *Migrator[T]) run(migrationType MigrationType, name string, f func(T)) (ok bool) {
	logger := m.ctx.Logger.With(slog.String("migration_type", string(migrationType)), slog.String("name", name))

	currentContext := m.ctx
	currentContext.Logger = logger
	currentContext.migrationType = migrationType

	tx, err := m.db.BeginTx(currentContext.Context, nil)
	if err != nil {
		m.ctx.Logger.Error("unable to start transaction", slog.String("error", err.Error()))
		return false
	}

	schema := m.schemaFactory(currentContext, tx)

	handleError := func(err any) {
		if err != nil {
			logger.Error("unable to run migration, rollback transaction...", slog.Any("error", err))

			err := tx.Rollback()
			if err != nil {
				logger.Error("unable to rollback transaction", slog.Any("error", err))
			}

			ok = false
		}
	}

	defer func() {
		if r := recover(); r != nil {
			handleError(r)
		}
	}()

	f(schema)

	if m.ctx.opts.DryRun {
		logger.Info("migration in dry run mode, rollback transaction...")
		err := tx.Rollback()
		if err != nil {
			logger.Error("unable to rollback transaction", slog.Any("error", err))
		}
		return true
	} else {
		err := tx.Commit()
		if err != nil {
			logger.Error("unable to commit transaction", slog.Any("error", err))
			return false
		}
	}

	return true
}

func (m *Migrator[T]) newSchema() T {
	return m.schemaFactory(m.ctx, m.db)
}

// Options returns a copy of the options.
func (m *Migrator[T]) Options() MigratorOption {
	return *m.ctx.opts
}

type ReversibleMigrationExec struct {
	migratorContext MigratorContext
}

type Directions struct {
	Up   func()
	Down func()
}

func (r *ReversibleMigrationExec) Reversible(directions Directions) {
	switch r.migratorContext.migrationType {
	case MigrationTypeUp:
		if directions.Up != nil {
			directions.Up()
		}
	case MigrationTypeDown:
		if directions.Down != nil {
			directions.Down()
		}
	}
}
