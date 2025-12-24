package amigo

import (
	"context"
	"fmt"
	"iter"
	"time"
)

// MigrationResult represents the result of applying a migration
type MigrationResult struct {
	Migration Migration
	Error     error
	Duration  time.Duration
}

// UpIterator returns an iterator that yields migration results as they are applied
func (r Runner) UpIterator(ctx context.Context, migrations []Migration, opts ...RunnerUpOptsFunc) iter.Seq[MigrationResult] {
	return func(yield func(MigrationResult) bool) {
		options := DefaultRunnerUpOpts()
		for _, opt := range opts {
			opt(&options)
		}

		if options.Steps == 0 {
			return
		}

		err := r.config.Driver.CreateSchemaMigrationsTableIfNotExists(ctx, r.config.DB)
		if err != nil {
			yield(MigrationResult{
				Error: fmt.Errorf("failed to create schema migrations table: %w", err),
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

		nonAppliedMigrations := r.filterNonAppliedMigrations(migrations, appliedMigrations)

		for _, m := range nonAppliedMigrations {
			start := time.Now()

			err := m.Up(ctx, r.config.DB)
			duration := time.Since(start)

			if err != nil {
				if !yield(MigrationResult{
					Migration: m,
					Error:     fmt.Errorf("failed to apply migration %s: %w", m.Name(), err),
					Duration:  duration,
				}) {
					return
				}
				return
			}

			record := MigrationRecord{
				Date: m.Date(),
				Name: m.Name(),
			}
			err = r.config.Driver.InsertMigrations(ctx, r.config.DB, []MigrationRecord{record})
			if err != nil {
				if !yield(MigrationResult{
					Migration: m,
					Error:     fmt.Errorf("failed to record applied migration %s: %w", m.Name(), err),
					Duration:  duration,
				}) {
					return
				}
				return
			}

			if !yield(MigrationResult{
				Migration: m,
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
