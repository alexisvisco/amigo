package schema

import (
	"fmt"
	"slices"

	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
)

func (m *Migrator[T]) detectMigrationsToExec(
	s Schema,
	migrationDirection types.MigrationDirection,
	allMigrations []Migration,
	version *string,
	steps *int, // only used for rollback
) (migrationsToApply []Migration, firstRun bool) {
	appliedVersions, err := utils.PanicToError1(s.FindAppliedVersions)
	if isTableDoesNotExists(err) {
		firstRun = true
		appliedVersions = []string{}
	} else if err != nil {
		m.migratorContext.RaiseError(err)
	}

	var versionsToApply []Migration
	var migrationsTimeFormat []string
	var versionToMigration = make(map[string]Migration)

	for _, migration := range allMigrations {
		migrationsTimeFormat = append(migrationsTimeFormat, migration.Date().UTC().Format(utils.FormatTime))
		versionToMigration[migrationsTimeFormat[len(migrationsTimeFormat)-1]] = migration
	}

	switch migrationDirection {
	case types.MigrationDirectionUp:
		if version != nil && *version != "" {
			if _, ok := versionToMigration[*version]; !ok {
				m.migratorContext.RaiseError(fmt.Errorf("version %s not found", *version))
			}

			if slices.Contains(appliedVersions, *version) {
				m.migratorContext.RaiseError(fmt.Errorf("version %s already applied", *version))
			}

			versionsToApply = append(versionsToApply, versionToMigration[*version])
			break
		}

		for _, currentMigrationVersion := range migrationsTimeFormat {
			if !slices.Contains(appliedVersions, currentMigrationVersion) {
				versionsToApply = append(versionsToApply, versionToMigration[currentMigrationVersion])
			}
		}
	case types.MigrationDirectionDown:
		if version != nil && *version != "" {
			if _, ok := versionToMigration[*version]; !ok {
				m.migratorContext.RaiseError(fmt.Errorf("version %s not found", *version))
			}

			if !slices.Contains(appliedVersions, *version) {
				m.migratorContext.RaiseError(fmt.Errorf("version %s not applied", *version))
			}

			versionsToApply = append(versionsToApply, versionToMigration[*version])
			break
		}

		step := 1
		if steps != nil && *steps > 0 {
			step = *steps
		}

		for i := len(allMigrations) - 1; i >= 0; i-- {
			if slices.Contains(appliedVersions, migrationsTimeFormat[i]) {
				versionsToApply = append(versionsToApply, versionToMigration[migrationsTimeFormat[i]])
			}

			if len(versionsToApply) == step {
				break
			}
		}
	}

	return versionsToApply, firstRun
}
