package schema

import "github.com/alexisvisco/amigo/pkg/types"

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
