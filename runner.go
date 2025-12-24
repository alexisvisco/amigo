package amigo

import (
	"slices"
	"sync"
)

type Runner struct {
	config                 Configuration
	ensureTableCreatedOnce sync.Once
}

func NewRunner(config Configuration) *Runner {
	return &Runner{config: config}
}

func (r *Runner) filterNonAppliedMigrations(migrations []Migration, appliedMigrations []MigrationRecord) []Migration {
	appliedDates := make(map[int64]struct{})
	for _, m := range appliedMigrations {
		appliedDates[m.Date] = struct{}{}
	}

	var result []Migration
	for _, m := range migrations {
		if _, applied := appliedDates[m.Date()]; !applied {
			result = append(result, m)
		}
	}

	// sort oldest first
	slices.SortFunc(result, func(a, b Migration) int {
		if a.Date() < b.Date() {
			return -1
		} else if a.Date() > b.Date() {
			return 1
		}
		return 0
	})

	return result
}

func (r *Runner) sortNewestFirstMigrationRecord(appliedMigrations []MigrationRecord) {
	slices.SortFunc(appliedMigrations, func(a, b MigrationRecord) int {
		if a.Date < b.Date {
			return 1
		} else if a.Date > b.Date {
			return -1
		}
		return 0
	})
}
