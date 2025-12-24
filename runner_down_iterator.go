package amigo

import (
	"context"
	"fmt"
	"iter"
	"time"
)

// DownIterator returns an iterator that yields migration results as they are reverted
func (r *Runner) DownIterator(ctx context.Context, migrations []Migration, opts ...RunnerDownOptsFunc) iter.Seq[MigrationResult] {
	return func(yield func(MigrationResult) bool) {
		options := defaultRunnerDownOpts()
		for _, opt := range opts {
			opt(&options)
		}

		if options.Steps == 0 {
			return
		}

		var tableErr error
		r.ensureTableCreatedOnce.Do(func() {
			tableErr = r.config.Driver.CreateSchemaMigrationsTableIfNotExists(ctx, r.config.DB)
		})
		if tableErr != nil {
			yield(MigrationResult{
				Error: fmt.Errorf("failed to create schema migrations table: %w", tableErr),
			})
			return
		}

		appliedMigrations, err := r.config.Driver.GetAppliedMigrations(ctx, r.config.DB)
		if err != nil {
			yield(MigrationResult{
				Error: fmt.Errorf("failed to get applied migrations: %w", err),
			})
			return
		}

		migrationsByDate := make(map[int64]Migration)
		for _, m := range migrations {
			migrationsByDate[m.Date()] = m
		}

		r.sortNewestFirstMigrationRecord(appliedMigrations)
		for _, am := range appliedMigrations {
			migration, exists := migrationsByDate[am.Date]
			if !exists {
				continue
			}

			start := time.Now()

			err := migration.Down(ctx, r.config.DB)
			duration := time.Since(start)

			if err != nil {
				if !yield(MigrationResult{
					Migration: migration,
					Error:     fmt.Errorf("failed to revert migration %s: %w", migration.Name(), err),
					Duration:  duration,
				}) {
					return
				}
				return
			}

			err = r.config.Driver.DeleteMigrations(ctx, r.config.DB, []int64{am.Date})
			if err != nil {
				if !yield(MigrationResult{
					Migration: migration,
					Error:     fmt.Errorf("failed to delete migration record %s: %w", migration.Name(), err),
					Duration:  duration,
				}) {
					return
				}
				return
			}

			if !yield(MigrationResult{
				Migration: migration,
				Error:     nil,
				Duration:  duration,
			}) {
				return
			}

			if options.Steps > 0 {
				options.Steps--
				if options.Steps == 0 {
					return
				}
			}
		}
	}
}
