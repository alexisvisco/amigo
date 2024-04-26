package schema

import (
	"context"
	"errors"
	"log/slog"
)

// MigratorContext is the context of the migrator.
// It contains the context, errors, options, and other useful information.
type MigratorContext struct {
	Context context.Context
	errors  error
	opts    *MigratorOption

	checkConstraintsCreated []CheckConstraintOptions
	foreignKeysCreated      []AddForeignKeyConstraintOptions
	indexesCreated          []IndexOptions
	extensionsCreated       []ExtensionOptions
	columnsCreated          []ColumnOptions
	primaryKeysCreated      []PrimaryKeyConstraintOptions

	extensionDropped []DropExtensionOptions

	reverseOperations []func()

	// Logger specifies the logger to use.
	Logger *slog.Logger

	migrationType MigrationType
}

// RaiseError will raise an error and stop the migration process if the `continue_on_error` option is not set.
// That's useful because each operation in the migration process is not a func(...) error, it's more convenient.
//
// If the ContinueOnError option is set, it will log the error and continue the migration process.
// If the error is a ForceStopError, it will stop the migration process even if the ContinueOnError option is set.
func (m *MigratorContext) RaiseError(err error) {
	m.addError(err)
	isForceStopError := errors.Is(err, &ForceStopError{})
	if !m.opts.ContinueOnError && !isForceStopError {
		panic(err)
	} else {
		m.Logger.Error("migration error found, continue due to `continue_on_error` option",
			slog.String("error", err.Error()))
	}
}

// Error returns the error of the migration process.
func (m *MigratorContext) Error() error {
	return m.errors
}

func (m *MigratorContext) addReverseOperation(f func()) {
	m.reverseOperations = append(m.reverseOperations, f)
}

func (m *MigratorContext) addError(err error) {
	if m.errors == nil {
		m.errors = err
		return
	}

	m.errors = errors.Join(m.errors, err)
}

func (m *MigratorContext) addCheckConstraintCreated(options CheckConstraintOptions) {
	m.checkConstraintsCreated = append(m.checkConstraintsCreated, options)
}

func (m *MigratorContext) addForeignKeyCreated(options AddForeignKeyConstraintOptions) {
	m.foreignKeysCreated = append(m.foreignKeysCreated, options)
}

func (m *MigratorContext) addIndexCreated(options IndexOptions) {
	m.indexesCreated = append(m.indexesCreated, options)
}

func (m *MigratorContext) addExtensionCreated(options ExtensionOptions) {
	m.extensionsCreated = append(m.extensionsCreated, options)
}

func (m *MigratorContext) addExtensionDropped(options DropExtensionOptions) {
	m.extensionDropped = append(m.extensionDropped, options)
}

func (m *MigratorContext) addColumnCreated(options ColumnOptions) {
	m.columnsCreated = append(m.columnsCreated, options)
}

func (m *MigratorContext) addPrimaryKeyCreated(options PrimaryKeyConstraintOptions) {
	m.primaryKeysCreated = append(m.primaryKeysCreated, options)
}
