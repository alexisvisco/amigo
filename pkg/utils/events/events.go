package events

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/utils"
	"time"
)

type EventName interface {
	EventName() string
}

type FileAddedEvent struct{ FileName string }

func (p FileAddedEvent) String() string { return fmt.Sprintf("+ file: %s", p.FileName) }

type FileModifiedEvent struct{ FileName string }

func (p FileModifiedEvent) String() string { return fmt.Sprintf("~ file: %s", p.FileName) }

type FolderAddedEvent struct{ FolderName string }

func (p FolderAddedEvent) String() string { return fmt.Sprintf("+ folder: %s", p.FolderName) }

type MessageEvent struct{ Message string }

func (p MessageEvent) String() string { return fmt.Sprintf("%s", p.Message) }

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
	MigrationVersion string
}

func (s SkipMigrationEvent) String() string {
	return fmt.Sprintf("------> skip migration: %s", s.MigrationVersion)
}

type SQLQueryEvent struct {
	Query string
}

func (s SQLQueryEvent) String() string {
	return fmt.Sprintf(s.Query)
}

type VersionAddedEvent struct {
	Version string
}

func (v VersionAddedEvent) String() string {
	return fmt.Sprintf("------> version migrated: %s", v.Version)
}

type VersionDeletedEvent struct {
	Version string
}

func (v VersionDeletedEvent) String() string {
	return fmt.Sprintf("------> version rolled back: %s", v.Version)
}
