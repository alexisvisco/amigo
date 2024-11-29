package schema

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/dblog"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
)

// MigratorOption is the option of the migrator.
type MigratorOption struct {
	// DryRun specifies if the migrator should perform the migrations without actually applying them.
	// Not compatible with TransactionNone.
	DryRun bool

	// ContinueOnError specifies if the migrator should continue running migrations even if an error occurs.
	// If you need to stop the migration process when an error occurs, wrap your error with NewForceStopError.
	ContinueOnError bool

	SchemaVersionTable TableName

	DBLogger dblog.DatabaseLogger

	// DumpSchemaFilePath is the path to the schema dump file.
	DumpSchemaFilePath *string

	// UseSchemaDump specifies if the migrator should use the schema. (if possible -> for fresh installation)
	UseSchemaDump bool
}

// Migration is the interface that describes a migration at is simplest form.
type Migration interface {
	Name() string
	Date() time.Time
}

// DetailedMigration is the interface that describes a migration with up and down operations.
type DetailedMigration[T Schema] interface {
	Up(T)
	Down(T)
	Name() string
	Date() time.Time
}

// SimpleMigration is the interface that describes a migration with a single operation.
type SimpleMigration[T Schema] interface {
	Change(T)
	Name() string
	Date() time.Time
}

type Factory[T Schema] func(ctx *MigratorContext, tx DB, db DB) T

// Migrator applies the migrations.
type Migrator[T Schema] struct {
	db              DBTX
	migratorContext *MigratorContext

	schemaFactory Factory[T]
	migrations    []func(T)
}

func (m *Migrator[T]) GetSchema() Schema {
	return m.schemaFactory(m.migratorContext, m.db, m.db)
}

// NewMigrator creates a new migrator.
func NewMigrator[T Schema](
	ctx context.Context,
	db DBTX,
	schemaFactory Factory[T],
	config *amigoconfig.Config,
) *Migrator[T] {
	return &Migrator[T]{
		db:            db,
		schemaFactory: schemaFactory,
		migratorContext: &MigratorContext{
			Context:         ctx,
			Config:          config,
			MigrationEvents: &MigrationEvents{},
		},
	}
}

func (m *Migrator[T]) Apply(direction types.MigrationDirection, version *string, steps *int, migrations []Migration) bool {
	db := m.schemaFactory(m.migratorContext, m.db, m.db)

	migrationsToExecute, firstRun := m.detectMigrationsToExec(
		db,
		direction,
		migrations,
		version,
		steps,
	)

	if len(migrationsToExecute) == 0 {
		logger.Info(events.MessageEvent{Message: "Found 0 migrations to apply"})
		return true
	}

	if firstRun && m.migratorContext.Config.Migration.UseSchemaDump {
		logger.Info(events.MessageEvent{Message: "We detect a fresh installation and applied the schema dump"})
		err := m.tryMigrateWithSchemaDump(migrationsToExecute)
		if err != nil {
			logger.Error(events.MessageEvent{Message: fmt.Sprintf("unable to apply schema dump: %v", err)})
			return false
		}

		logger.Info(events.MessageEvent{Message: "Schema dump applied successfully"})
		return true
	}

	m.ToggleDBLog(true)
	defer m.ToggleDBLog(false)

	for _, migration := range migrationsToExecute {
		var migrationFunc func(T)

		switch t := migration.(type) {
		case DetailedMigration[T]:
			switch direction {
			case types.MigrationDirectionUp:
				migrationFunc = t.Up
			case types.MigrationDirectionDown, types.MigrationDirectionNotReversible:
				direction = types.MigrationDirectionNotReversible
				migrationFunc = t.Down
			}
		case SimpleMigration[T]:
			migrationFunc = t.Change
		default:
			reflectionType := fmt.Sprintf("%T", migration)
			driverValue := fmt.Sprintf("%v", reflect.TypeOf(new(T)).Elem())

			logger.Error(events.MessageEvent{Message: fmt.Sprintf("Migration %s is not a valid migration type, found %s (driver is %s). Did you set DSN correctly ?",
				migration.Name(), reflectionType, driverValue)})
			return false
		}

		switch direction {
		case types.MigrationDirectionUp:
			logger.Info(events.MigrateUpEvent{MigrationName: migration.Name(), Time: migration.Date()})
		case types.MigrationDirectionDown, types.MigrationDirectionNotReversible:
			logger.Info(events.MigrateDownEvent{MigrationName: migration.Name(), Time: migration.Date()})
		}

		if migrationFunc == nil {
			logger.Error(events.MessageEvent{Message: fmt.Sprintf("Migration %s is not a valid migration",
				migration.Name())})
			return false
		}

		if !m.run(direction, fmt.Sprint(migration.Date().UTC().Format(utils.FormatTime)), migrationFunc) {
			return false
		}
	}

	return true
}

func (m *Migrator[T]) NewSchema() T {
	return m.schemaFactory(m.migratorContext, m.db, m.db)
}

func (m *Migrator[T]) ToggleDBLog(b bool) {
	// todo: adjust logger
}
