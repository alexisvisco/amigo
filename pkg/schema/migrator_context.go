package schema

import (
	"context"
	"errors"
	"fmt"
	"github.com/alexisvisco/mig/pkg/types"
	"github.com/alexisvisco/mig/pkg/utils/tracker"
)

// MigratorContext is the context of the migrator.
// It contains the context, errors, options, and other useful information.
type MigratorContext struct {
	Context         context.Context
	errors          error
	MigratorOptions *MigratorOption
	MigrationEvents *MigrationEvents

	Track tracker.Tracker

	MigrationDirection types.MigrationDirection
}

type MigrationEvents struct {
	checkConstraintsCreated []CheckConstraintOptions
	foreignKeysCreated      []AddForeignKeyConstraintOptions
	indexesCreated          []IndexOptions
	extensionsCreated       []ExtensionOptions
	columnsCreated          []ColumnOptions
	primaryKeysCreated      []PrimaryKeyConstraintOptions
	tablesCreated           []TableOptions

	columnsRenamed []RenameColumnOptions
	columnComments []ColumnCommentOptions

	extensionsDropped            []DropExtensionOptions
	tablesDropped                []DropTableOptions
	indexesDropped               []DropIndexOptions
	columnsDropped               []DropColumnOptions
	checkConstraintsDropped      []DropCheckConstraintOptions
	foreignKeyConstraintsDropped []DropForeignKeyConstraintOptions
	primaryKeyConstraintsDropped []DropPrimaryKeyConstraintOptions
}

// ForceStopError is an error that stops the migration process even if the `continue_on_error` option is set.
type ForceStopError struct{ error }

// NewForceStopError creates a new ForceStopError.
func NewForceStopError(err error) *ForceStopError {
	return &ForceStopError{err}
}

// RaiseError will raise an error and stop the migration process if the `continue_on_error` option is not set.
// That's useful because each operation in the migration process is not a func(...) error, it's more convenient.
//
// If the ContinueOnError option is set, it will log the error and continue the migration process.
// If the error is a ForceStopError, it will stop the migration process even if the ContinueOnError option is set.
func (m *MigratorContext) RaiseError(err error) {
	m.addError(err)
	isForceStopError := errors.Is(err, &ForceStopError{})
	if !m.MigratorOptions.ContinueOnError && !isForceStopError {
		panic(err)
	} else {
		m.Track.AddEvent(tracker.InfoEvent{
			Message: fmt.Sprintf("migration error found, continue due to `continue_on_error` option: %s", err.Error()),
		})
	}
}

// Error returns the error of the migration process.
func (m *MigratorContext) Error() error {
	return m.errors
}

func (m *MigratorContext) addError(err error) {
	if m.errors == nil {
		m.errors = err
		return
	}

	m.errors = errors.Join(m.errors, err)
}

func (m *MigratorContext) AddCheckConstraintCreated(options CheckConstraintOptions) {
	m.MigrationEvents.checkConstraintsCreated = append(m.MigrationEvents.checkConstraintsCreated, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddForeignKeyCreated(options AddForeignKeyConstraintOptions) {
	m.MigrationEvents.foreignKeysCreated = append(m.MigrationEvents.foreignKeysCreated, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddIndexCreated(options IndexOptions) {
	m.MigrationEvents.indexesCreated = append(m.MigrationEvents.indexesCreated, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddExtensionCreated(options ExtensionOptions) {
	m.MigrationEvents.extensionsCreated = append(m.MigrationEvents.extensionsCreated, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddExtensionDropped(options DropExtensionOptions) {
	m.MigrationEvents.extensionsDropped = append(m.MigrationEvents.extensionsDropped, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddColumnCreated(options ColumnOptions) {
	m.MigrationEvents.columnsCreated = append(m.MigrationEvents.columnsCreated, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddPrimaryKeyCreated(options PrimaryKeyConstraintOptions) {
	m.MigrationEvents.primaryKeysCreated = append(m.MigrationEvents.primaryKeysCreated, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddTableCreated(options TableOptions) {
	m.MigrationEvents.tablesCreated = append(m.MigrationEvents.tablesCreated, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddTableDropped(options DropTableOptions) {
	m.MigrationEvents.tablesDropped = append(m.MigrationEvents.tablesDropped, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddIndexDropped(options DropIndexOptions) {
	m.MigrationEvents.indexesDropped = append(m.MigrationEvents.indexesDropped, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddRenameColumn(options RenameColumnOptions) {
	m.MigrationEvents.columnsRenamed = append(m.MigrationEvents.columnsRenamed, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddColumnDropped(options DropColumnOptions) {
	m.MigrationEvents.columnsDropped = append(m.MigrationEvents.columnsDropped, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddColumnComment(options ColumnCommentOptions) {
	m.MigrationEvents.columnComments = append(m.MigrationEvents.columnComments, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddCheckConstraintDropped(options DropCheckConstraintOptions) {
	m.MigrationEvents.checkConstraintsDropped = append(m.MigrationEvents.checkConstraintsDropped, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddForeignKeyConstraintDropped(options DropForeignKeyConstraintOptions) {
	m.MigrationEvents.foreignKeyConstraintsDropped = append(m.MigrationEvents.foreignKeyConstraintsDropped, options)
	m.Track.AddEvent(options)
}

func (m *MigratorContext) AddPrimaryKeyConstraintDropped(options DropPrimaryKeyConstraintOptions) {
	m.MigrationEvents.primaryKeyConstraintsDropped = append(m.MigrationEvents.primaryKeyConstraintsDropped, options)
	m.Track.AddEvent(options)
}
