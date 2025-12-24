package amigo

import (
	"context"
	"fmt"
	"slices"
)

func (r Runner) GetMigrationsStatuses(ctx context.Context, migrations []Migration) (all []MigrationStatus, err error) {
	err = r.config.Driver.CreateSchemaMigrationsTableIfNotExists(ctx, r.config.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema migrations table: %w", err)
	}

	appliedMigrations, err := r.config.Driver.GetAppliedMigrations(ctx, r.config.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[int64]MigrationRecord)
	for _, am := range appliedMigrations {
		appliedMap[am.Date] = am
	}

	for _, m := range migrations {
		status := MigrationStatus{
			Migration: MigrationRecord{
				Date: m.Date(),
				Name: m.Name(),
			},
		}

		if applied, exists := appliedMap[m.Date()]; exists {
			status.Applied = true
			status.Migration.AppliedAt = applied.AppliedAt
		}

		all = append(all, status)
	}

	// sort all migrations by date ascending = oldest first (chronological order)
	slices.SortFunc(all, func(a, b MigrationStatus) int {
		if a.Migration.Date < b.Migration.Date {
			return -1
		} else if a.Migration.Date > b.Migration.Date {
			return 1
		}
		return 0
	})

	return all, nil
}
