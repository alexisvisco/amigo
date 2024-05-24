package schema

import "github.com/alexisvisco/amigo/pkg/types"

// ReversibleMigrationExecutor is a helper to execute reversible migrations in a change method.
// Since you may have custom code in it, you must provide a way to up and down the user defined code.
type ReversibleMigrationExecutor struct {
	migratorContext *MigratorContext
}

func NewReversibleMigrationExecutor(ctx *MigratorContext) *ReversibleMigrationExecutor {
	return &ReversibleMigrationExecutor{migratorContext: ctx}
}

type Directions struct {
	Up   func()
	Down func()
}

func (r *ReversibleMigrationExecutor) Reversible(directions Directions) {
	currentMigrationDirection := r.migratorContext.MigrationDirection
	r.migratorContext.MigrationDirection = types.MigrationDirectionNotReversible
	defer func() {
		r.migratorContext.MigrationDirection = currentMigrationDirection
	}()

	switch currentMigrationDirection {
	case types.MigrationDirectionUp:
		if directions.Up != nil {
			directions.Up()
		}
	case types.MigrationDirectionDown:
		if directions.Down != nil {
			directions.Down()
		}
	}
}
