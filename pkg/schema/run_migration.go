package schema

import (
	"fmt"

	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
)

// run runs the migration.
func (m *Migrator[T]) run(migrationType types.MigrationDirection, version string, f func(T)) (ok bool) {
	currentContext := m.migratorContext
	currentContext.MigrationDirection = migrationType

	tx, err := m.db.BeginTx(currentContext.Context, nil)
	if err != nil {
		logger.Error(events.MessageEvent{Message: "unable to start transaction"})
		return false
	}

	schema := m.schemaFactory(currentContext, tx, m.db)

	handleError := func(err any) {
		if err != nil {
			logger.Error(events.MessageEvent{Message: fmt.Sprintf("migration failed, rollback due to: %v", err)})

			err := tx.Rollback()
			if err != nil {
				logger.Error(events.MessageEvent{Message: "unable to rollback transaction"})
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
	case types.MigrationDirectionDown, types.MigrationDirectionNotReversible:
		schema.RemoveVersion(version)
	}

	if m.migratorContext.Config.Migration.DryRun {
		logger.Info(events.MessageEvent{Message: "migration in dry run mode, rollback transaction..."})
		err := tx.Rollback()
		if err != nil {
			logger.Error(events.MessageEvent{Message: "unable to rollback transaction"})
		}
		return true
	} else {
		err := tx.Commit()
		if err != nil {
			logger.Error(events.MessageEvent{Message: "unable to commit transaction"})
			return false
		}
	}

	return true
}
