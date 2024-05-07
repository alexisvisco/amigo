package schema

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/alexisvisco/mig/pkg/types"
	"github.com/alexisvisco/mig/pkg/utils"
	"github.com/alexisvisco/mig/pkg/utils/tracker"
	"slices"
	"time"
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

type SchemaFactory[T Schema] func(*MigratorContext, DB) T

// Migrator applies the migrations.
type Migrator[T Schema] struct {
	db  *sql.DB
	ctx *MigratorContext

	schemaFactory SchemaFactory[T]
	migrations    []func(T)
}

// NewMigrator creates a new migrator.
func NewMigrator[T Schema](
	ctx context.Context,
	db *sql.DB,
	tracker tracker.Tracker,
	schemaFactory SchemaFactory[T],
	opts *MigratorOption,
) *Migrator[T] {
	return &Migrator[T]{
		db:            db,
		schemaFactory: schemaFactory,
		ctx: &MigratorContext{
			Context:         ctx,
			MigratorOptions: opts,
			MigrationEvents: &MigrationEvents{},
			Track:           tracker,
		},
	}
}

func (m *Migrator[T]) Apply(direction types.MigrationDirection, version *string, steps *int, migrations []Migration) bool {
	db := m.schemaFactory(m.ctx, m.db)

	migrationsToExecute := make([]Migration, 0, len(migrations))
	if !db.TableExist(m.Options().SchemaVersionTable) {
		// the first migration is always the creation of the schema version table
		migrationsToExecute = append(migrationsToExecute, migrations[0])
	} else {
		migrationsToExecute = m.deduceMigrationsToExecute(db,
			direction,
			migrations,
			version,
			steps)
	}

	if len(migrationsToExecute) == 0 {
		m.ctx.Track.AddEvent(tracker.InfoEvent{Message: "no migrations to apply"})
		return true
	}

	for _, migration := range migrationsToExecute {
		var migrationFunc func(T)

		switch t := migration.(type) {
		case DetailedMigration[T]:
			switch direction {
			case types.MigrationDirectionUp:
				migrationFunc = t.Up
			case types.MigrationDirectionDown:
				direction = types.MigrationDirectionNotReversible
				migrationFunc = t.Down
			}
		case SimpleMigration[T]:
			migrationFunc = t.Change
		default:
			m.ctx.RaiseError(errors.New("invalid migration type"))
		}

		switch direction {
		case types.MigrationDirectionUp:
			m.ctx.Track.AddEvent(tracker.MigrateUpEvent{MigrationName: migration.Name(), Time: migration.Date()})
		case types.MigrationDirectionDown:
			m.ctx.Track.AddEvent(tracker.MigrateDownEvent{MigrationName: migration.Name(), Time: migration.Date()})
		}

		if !m.run(direction, fmt.Sprint(migration.Date().UTC().Format(utils.FormatTime)), migrationFunc) {
			return false
		}
	}

	return true
}

func (m *Migrator[T]) deduceMigrationsToExecute(
	s Schema,
	migtype types.MigrationDirection,
	migrations []Migration,
	version *string,
	steps *int, // only used for rollback
) []Migration {
	appliedVersions := s.FindAppliedVersions()
	var versionsToApply []Migration
	var migrationsTimeFormat []string
	var versionToMigration = make(map[string]Migration)

	for _, migration := range migrations {
		migrationsTimeFormat = append(migrationsTimeFormat, migration.Date().UTC().Format(utils.FormatTime))
		versionToMigration[migrationsTimeFormat[len(migrationsTimeFormat)-1]] = migration
	}

	switch migtype {
	case types.MigrationDirectionUp:
		if version != nil && *version != "" {
			if versionToMigration[*version] == nil {
				m.ctx.RaiseError(fmt.Errorf("version %s not found", *version))
			}

			if slices.Contains(appliedVersions, *version) {
				m.ctx.RaiseError(fmt.Errorf("version %s already applied", *version))
			}

			versionsToApply = append(versionsToApply, versionToMigration[*version])
			break
		}

		fmt.Println(appliedVersions)
		fmt.Println(migrationsTimeFormat)

		for _, currentMigrationVersion := range migrationsTimeFormat {
			if !slices.Contains(appliedVersions, currentMigrationVersion) {
				versionsToApply = append(versionsToApply, versionToMigration[currentMigrationVersion])
			}
		}
	case types.MigrationDirectionDown:
		if version != nil && *version != "" {
			if versionToMigration[*version] == nil {
				m.ctx.RaiseError(fmt.Errorf("version %s not found", *version))
			}

			if !slices.Contains(appliedVersions, *version) {
				m.ctx.RaiseError(fmt.Errorf("version %s not applied", *version))
			}

			versionsToApply = append(versionsToApply, versionToMigration[*version])
			break
		}

		step := 1
		if steps != nil && *steps > 0 {
			step = *steps
		}

		for i := len(migrations) - 1; i >= 0; i-- {
			if slices.Contains(appliedVersions, migrationsTimeFormat[i]) {
				versionsToApply = append(versionsToApply, versionToMigration[migrationsTimeFormat[i]])
			}

			if len(versionsToApply) == step {
				break
			}
		}
	}

	return versionsToApply
}

// run runs the migration.
func (m *Migrator[T]) run(migrationType types.MigrationDirection, version string, f func(T)) (ok bool) {

	currentContext := m.ctx
	currentContext.MigrationDirection = migrationType

	tx, err := m.db.BeginTx(currentContext.Context, nil)
	if err != nil {
		m.ctx.Track.AddEvent(tracker.InfoEvent{Message: "unable to start transaction"})
		return false
	}

	schema := m.schemaFactory(currentContext, tx)

	handleError := func(err any) {
		if err != nil {
			m.ctx.Track.AddEvent(tracker.InfoEvent{Message: fmt.Sprintf("migration failed, rollback due to: %v", err)})

			err := tx.Rollback()
			if err != nil {
				m.ctx.Track.AddEvent(tracker.InfoEvent{Message: "unable to rollback transaction"})
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

	switch migrationType {
	case types.MigrationDirectionUp:
		schema.AddVersion(version)
	case types.MigrationDirectionDown:
		schema.RemoveVersion(version)
	}

	if m.ctx.MigratorOptions.DryRun {
		m.ctx.Track.AddEvent(tracker.InfoEvent{Message: "migration in dry run mode, rollback transaction..."})
		err := tx.Rollback()
		if err != nil {
			m.ctx.Track.AddEvent(tracker.InfoEvent{Message: "unable to rollback transaction"})
		}
		return true
	} else {
		err := tx.Commit()
		if err != nil {
			m.ctx.Track.AddEvent(tracker.InfoEvent{Message: "unable to commit transaction"})
			return false
		}
	}

	return true
}

func (m *Migrator[T]) NewSchema() T {
	return m.schemaFactory(m.ctx, m.db)
}

// Options returns a copy of the options.
func (m *Migrator[T]) Options() MigratorOption {
	return *m.ctx.MigratorOptions
}
