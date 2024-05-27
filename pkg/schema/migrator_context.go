package schema

import (
	"context"
	"errors"
	"fmt"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
)

// MigratorContext is the context of the migrator.
// It contains the context, errors, options, and other useful information.
type MigratorContext struct {
	Context         context.Context
	errors          error
	MigratorOptions *MigratorOption
	MigrationEvents *MigrationEvents

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
	versionCreated          []string
	enumCreated             []CreateEnumOptions
	enumValueCreated        []AddEnumValueOptions

	columnsRenamed       []RenameColumnOptions
	tablesRenamed        []RenameTableOptions
	changeColumnTypes    []ChangeColumnTypeOptions
	columnComments       []ColumnCommentOptions
	tableComments        []TableCommentOptions
	changeColumnDefaults []ChangeColumnDefaultOptions
	renameEnums          []RenameEnumOptions
	renameEnumValues     []RenameEnumValueOptions

	extensionsDropped            []DropExtensionOptions
	tablesDropped                []DropTableOptions
	indexesDropped               []DropIndexOptions
	columnsDropped               []DropColumnOptions
	checkConstraintsDropped      []DropCheckConstraintOptions
	foreignKeyConstraintsDropped []DropForeignKeyConstraintOptions
	primaryKeyConstraintsDropped []DropPrimaryKeyConstraintOptions
	versionDeleted               []string
	enumDropped                  []DropEnumOptions
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
		logger.Info(events.MessageEvent{
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
	logger.Info(options)
}

func (m *MigratorContext) AddForeignKeyCreated(options AddForeignKeyConstraintOptions) {
	m.MigrationEvents.foreignKeysCreated = append(m.MigrationEvents.foreignKeysCreated, options)
	logger.Info(options)
}

func (m *MigratorContext) AddIndexCreated(options IndexOptions) {
	m.MigrationEvents.indexesCreated = append(m.MigrationEvents.indexesCreated, options)
	logger.Info(options)
}

func (m *MigratorContext) AddExtensionCreated(options ExtensionOptions) {
	m.MigrationEvents.extensionsCreated = append(m.MigrationEvents.extensionsCreated, options)
	logger.Info(options)
}

func (m *MigratorContext) AddExtensionDropped(options DropExtensionOptions) {
	m.MigrationEvents.extensionsDropped = append(m.MigrationEvents.extensionsDropped, options)
	logger.Info(options)
}

func (m *MigratorContext) AddColumnCreated(options ColumnOptions) {
	m.MigrationEvents.columnsCreated = append(m.MigrationEvents.columnsCreated, options)
	logger.Info(options)
}

func (m *MigratorContext) AddPrimaryKeyCreated(options PrimaryKeyConstraintOptions) {
	m.MigrationEvents.primaryKeysCreated = append(m.MigrationEvents.primaryKeysCreated, options)
	logger.Info(options)
}

func (m *MigratorContext) AddTableCreated(options TableOptions) {
	m.MigrationEvents.tablesCreated = append(m.MigrationEvents.tablesCreated, options)
	logger.Info(options)
}

func (m *MigratorContext) AddTableDropped(options DropTableOptions) {
	m.MigrationEvents.tablesDropped = append(m.MigrationEvents.tablesDropped, options)
	logger.Info(options)
}

func (m *MigratorContext) AddIndexDropped(options DropIndexOptions) {
	m.MigrationEvents.indexesDropped = append(m.MigrationEvents.indexesDropped, options)
	logger.Info(options)
}

func (m *MigratorContext) AddRenameColumn(options RenameColumnOptions) {
	m.MigrationEvents.columnsRenamed = append(m.MigrationEvents.columnsRenamed, options)
	logger.Info(options)
}

func (m *MigratorContext) AddColumnDropped(options DropColumnOptions) {
	m.MigrationEvents.columnsDropped = append(m.MigrationEvents.columnsDropped, options)
	logger.Info(options)
}

func (m *MigratorContext) AddColumnComment(options ColumnCommentOptions) {
	m.MigrationEvents.columnComments = append(m.MigrationEvents.columnComments, options)
	logger.Info(options)
}

func (m *MigratorContext) AddCheckConstraintDropped(options DropCheckConstraintOptions) {
	m.MigrationEvents.checkConstraintsDropped = append(m.MigrationEvents.checkConstraintsDropped, options)
	logger.Info(options)
}

func (m *MigratorContext) AddForeignKeyConstraintDropped(options DropForeignKeyConstraintOptions) {
	m.MigrationEvents.foreignKeyConstraintsDropped = append(m.MigrationEvents.foreignKeyConstraintsDropped, options)
	logger.Info(options)
}

func (m *MigratorContext) AddPrimaryKeyConstraintDropped(options DropPrimaryKeyConstraintOptions) {
	m.MigrationEvents.primaryKeyConstraintsDropped = append(m.MigrationEvents.primaryKeyConstraintsDropped, options)
	logger.Info(options)
}

func (m *MigratorContext) AddVersionCreated(version string) {
	m.MigrationEvents.versionCreated = append(m.MigrationEvents.versionCreated, version)
	logger.Info(events.VersionAddedEvent{Version: version})
}

func (m *MigratorContext) AddVersionDeleted(version string) {
	m.MigrationEvents.versionDeleted = append(m.MigrationEvents.versionDeleted, version)
	logger.Info(events.VersionDeletedEvent{Version: version})
}

func (m *MigratorContext) AddTableRenamed(options RenameTableOptions) {
	m.MigrationEvents.tablesRenamed = append(m.MigrationEvents.tablesRenamed, options)
	logger.Info(options)
}

func (m *MigratorContext) AddChangeColumnType(options ChangeColumnTypeOptions) {
	m.MigrationEvents.changeColumnTypes = append(m.MigrationEvents.changeColumnTypes, options)
	logger.Info(options)
}

func (m *MigratorContext) AddChangeColumnDefault(options ChangeColumnDefaultOptions) {
	m.MigrationEvents.changeColumnDefaults = append(m.MigrationEvents.changeColumnDefaults, options)
	logger.Info(options)
}

func (m *MigratorContext) AddTableComment(options TableCommentOptions) {
	m.MigrationEvents.tableComments = append(m.MigrationEvents.tableComments, options)
	logger.Info(options)
}

func (m *MigratorContext) AddEnumCreated(options CreateEnumOptions) {
	m.MigrationEvents.enumCreated = append(m.MigrationEvents.enumCreated, options)
	logger.Info(options)
}

func (m *MigratorContext) AddEnumDropped(options DropEnumOptions) {
	m.MigrationEvents.enumDropped = append(m.MigrationEvents.enumDropped, options)
	logger.Info(options)
}

func (m *MigratorContext) AddEnumValueCreated(options AddEnumValueOptions) {
	m.MigrationEvents.enumValueCreated = append(m.MigrationEvents.enumValueCreated, options)
	logger.Info(options)
}

func (m *MigratorContext) AddRenameEnum(options RenameEnumOptions) {
	m.MigrationEvents.renameEnums = append(m.MigrationEvents.renameEnums, options)
	logger.Info(options)
}

func (m *MigratorContext) AddRenameEnumValue(options RenameEnumValueOptions) {
	m.MigrationEvents.renameEnumValues = append(m.MigrationEvents.renameEnumValues, options)
	logger.Info(options)
}
