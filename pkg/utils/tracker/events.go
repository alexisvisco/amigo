package tracker

import (
	"fmt"
	"github.com/alexisvisco/mig/pkg/utils"
	"time"
)

type FileAddedEvent struct{ FileName string }

func (p FileAddedEvent) String() string { return fmt.Sprintf("+ file: %s", p.FileName) }

type FileModifiedEvent struct{ FileName string }

func (p FileModifiedEvent) String() string { return fmt.Sprintf("~ file: %s", p.FileName) }

type FolderAddedEvent struct{ FolderName string }

func (p FolderAddedEvent) String() string { return fmt.Sprintf("+ folder: %s", p.FolderName) }

type InfoEvent struct{ Message string }

func (p InfoEvent) String() string { return fmt.Sprintf("%s", p.Message) }

type RawEvent struct{ Message string }

func (p RawEvent) String() string { return p.Message }

type MeasurementEvent struct{ TimeElapsed time.Duration }

func (m MeasurementEvent) String() string { return fmt.Sprintf("  done in %s", m.TimeElapsed) }

type MigrateUpEvent struct {
	MigrationName string
	Time          time.Time
}

func (m MigrateUpEvent) String() string {
	return fmt.Sprintf("------> migrating: %s version: %s", m.MigrationName, m.Time.Format(utils.FormatTime))
}

type MigrateDownEvent struct {
	MigrationName string
	Time          time.Time
}

func (m MigrateDownEvent) String() string {
	return fmt.Sprintf("------> rollback: %s version: %s", m.MigrationName, m.Time.Format(utils.FormatTime))
}

type SkipMigrationEvent struct {
	MigrationVersion int64
}

func (s SkipMigrationEvent) String() string {
	return fmt.Sprintf("------> skip migration: %d", s.MigrationVersion)
}
