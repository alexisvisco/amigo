package schema

import (
	"fmt"
	"strings"
)

type TableName string

// Schema returns the schema part of the table name.
func (t TableName) Schema() string {
	index := strings.IndexByte(string(t), '.')
	if index != -1 {
		return string(t[:index])
	}
	return "public"
}

// HasSchema returns if the table name has a schema.
func (t TableName) HasSchema() bool {
	return strings.Contains(string(t), ".")
}

// Name returns the name part of the table name.
func (t TableName) Name() string {
	index := strings.IndexByte(string(t), '.')
	if index != -1 {
		return string(t[index+1:])
	}
	return string(t)
}

// String returns the string representation of the table name.
func (t TableName) String() string {
	return string(t)
}

// Table returns a new TableName with the schema and table name.
// But you can also use regular string, this function is when you have dynamic schema names. (like in the tests)
func Table(name string, schema ...string) TableName {
	if len(schema) == 0 {
		return TableName(name)
	}
	return TableName(fmt.Sprintf("%s.%s", schema[0], name))
}

type ConstraintOption interface {
	ConstraintType() string
}

var DefaultCheckConstraintNameBuilder = func(table TableName, constraintName string) string {
	return fmt.Sprintf("ck_%s_%s", table.Name(), constraintName)
}

type CheckConstraintOptions struct {
	Table          TableName
	Expression     string
	ConstraintName string

	// ConstraintNameBuilder will build the name of the constraint. If nil, a default name will be used.
	// By default, is nil.
	ConstraintNameBuilder func(table TableName, constraintName string) string

	// IfNotExists Silently ignore if the constraint already exists, rather than raise an error.
	IfNotExists bool

	// Postgres specific  ----------------------------

	// Validate Specify whether the constraint should be validated. Defaults to true.
	Validate *bool

	// End of Postgres specific ----------------------------
}

// ConstraintType returns the type of the constraint.
func (c CheckConstraintOptions) ConstraintType() string {
	return "check"
}

// BuildConstraintName returns the name of the constraint.
// If ConstraintNameBuilder is set, it will be used to build the name.
// Default name is `ck_{table_name}_{constraint_name}`
func (c CheckConstraintOptions) BuildConstraintName(table TableName, constraintName string) string {
	if c.ConstraintName != "" {
		return c.ConstraintName
	}

	if c.ConstraintNameBuilder == nil {
		return DefaultCheckConstraintNameBuilder(table, constraintName)
	}

	return c.ConstraintNameBuilder(table, constraintName)
}

var DefaultForeignKeyNameBuilder = func(fromTable TableName, toTable TableName) string {
	return fmt.Sprintf("fk_%s_%s", fromTable.Name(), toTable.Name())
}

type AddForeignKeyConstraintOptions struct {
	FromTable      TableName
	ToTable        TableName
	ForeignKeyName string

	// ForeignKeyNameBuilder will build the name of the foreign key. If nil, a default name will be used.
	ForeignKeyNameBuilder func(fromTable TableName, toTable TableName) string

	// Column is the foreign key column name on FromTable. Defaults to ToTable.Singularize + "_id". Pass an array to
	// create a composite foreign key.
	Column string

	// PrimaryKey is the primary key column name on ToTable. Defaults to "id".
	// foreign key.
	PrimaryKey string

	// CompositePrimaryKey is the primary key column names on ToTable.
	CompositePrimaryKey []string

	// OnDelete is the action that happens ON DELETE. Valid values are "nullify", "cascade", and "restrict" or custom action.
	OnDelete string

	// OnUpdate is the action that happens ON UPDATE. Valid values are "nullify", "cascade", and "restrict" or custom action.
	OnUpdate string

	// IfNotExists specifies if the foreign key already exists to not try to re-add it. This will avoid duplicate column errors.
	IfNotExists bool

	// Postgres specific ----------------------------

	// Validate specifies whether the constraint should be validated. Defaults to true.
	Validate *bool

	// Deferrable specifies whether the foreign key should be deferrable.
	// Could be DEFERRABLE, NOT DEFERRABLE, INITIALLY DEFERRED, INITIALLY IMMEDIATE or both.
	Deferrable string

	// End of Postgres specific ----------------------------
}

// ConstraintType returns the type of the constraint.
func (f AddForeignKeyConstraintOptions) ConstraintType() string {
	return "foreign_key"
}

// BuildForeignKeyName returns the name of the foreign key.
// If ForeignKeyNameBuilder is set, it will be used to build the name.
// Default name is `fk_{from_table}_{to_table}`
func (f AddForeignKeyConstraintOptions) BuildForeignKeyName(fromTable TableName, toTable TableName) string {
	if f.ForeignKeyName != "" {
		return f.ForeignKeyName
	}

	if f.ForeignKeyNameBuilder == nil {
		return DefaultForeignKeyNameBuilder(fromTable, toTable)
	}

	return f.ForeignKeyNameBuilder(fromTable, toTable)
}

var DefaultIndexNameBuilder = func(table TableName, columns []string) string {
	return fmt.Sprintf("idx_%s_%s", table.Name(), strings.Join(columns, "_"))
}

type IndexOptions struct {
	Table     TableName
	Columns   []string
	IndexName string

	// IndexNameBuilder will build the name of the index. If nil, a default name will be used.
	IndexNameBuilder func(table TableName, columns []string) string

	// IfNotExists Silently ignore if the index already exists, rather than raise an error.
	IfNotExists bool

	// Unique specifies if the index should be unique.
	Unique bool

	// Concurrent specifies if the index should be created concurrently.
	Concurrent bool

	// Method specifies the index method.
	Method string

	// Order specifies the order of the index.
	Order string

	// OrderPerColumn specifies the order of the index per column.
	OrderPerColumn map[string]string

	// Postgres specific ----------------------------

	// Predicate specifies the predicate for the index.
	Predicate string

	// End Postgres specific ----------------------------

}

// BuildIndexName returns the name of the index.
// If IndexNameBuilder is set, it will be used to build the name.
// Default name is `idx_{table_name}_{Columns}`
func (i IndexOptions) BuildIndexName(table TableName, columns []string) string {
	if i.IndexName != "" {
		return i.IndexName
	}

	if i.IndexNameBuilder == nil {
		return DefaultIndexNameBuilder(table, columns)
	}

	return i.IndexNameBuilder(table, columns)
}

type DropIndexOptions struct {
	Table TableName

	Columns []string

	IndexName string

	IndexNameBuilder func(table TableName, columns []string) string

	// IfExists add IF EXISTS to the query.
	IfExists bool

	// Reversible will allow the migrator to reverse the operation by creating the index.
	Reversible *IndexOptions
}

func (d DropIndexOptions) BuildIndexName(table TableName, columns []string) string {
	if d.IndexName != "" {
		return d.IndexName
	}

	if d.IndexNameBuilder == nil {
		return DefaultIndexNameBuilder(table, columns)
	}

	return d.IndexNameBuilder(table, columns)
}

type ExtensionOptions struct {
	ExtensionName string

	// Schema is the schema where the extension will be created.
	Schema string

	// IfNotExists add IF NOT EXISTS to the query.
	IfNotExists bool
}

type DropExtensionOptions struct {
	ExtensionName string

	// IfExists add IF EXISTS to the query.
	IfExists bool

	// Reversible will allow the migrator to reverse the operation by creating the extension.
	Reversible *ExtensionOptions
}

type ColumnOptions struct {
	Table      TableName
	ColumnName string

	// PrimaryKey specifies if the column is a primary key.
	PrimaryKey bool

	// Precision is the precision of the column.
	// Mainly supported for the ColumnTypeDecimal, ColumnTypeNumeric, ColumnTypeDatetime, and ColumnTypeTime
	// The precision is the number of significant digits in a number.
	Precision int

	// Scale is the scale of the column.
	// Mainly supported for Specifies the scale for the ColumnTypeDecimal, ColumnTypeNumeric
	// The scale is the number of digits to the right of the decimal point in a number.
	Scale int

	// ColumnType is the type of the column.
	// Could be a custom one but it's recommended to use the predefined ones for portability.
	ColumnType ColumnType

	// Default is the default value of the column.
	Default string

	// NotNull specifies if the column can be null.
	NotNull bool

	// Limit is a maximum column length. This is the number of characters for a ColumnTypeString column and number of
	// bytes for ColumnTypeText, ColumnTypeBinary, ColumnTypeBlob, and ColumnTypeInteger Columns.
	Limit int

	// IfNotExists specifies if the column already exists to not try to re-add it. This will avoid duplicate column errors.
	IfNotExists bool

	// Array specifies if the column is an array.
	Array bool

	// Comment is the Comment of the column.
	Comment string

	Constraints []ConstraintOption
}

type ColumnCommentOptions struct {
	Table TableName

	// ColumnName is the name of the column.
	ColumnName string

	// Comment is the Comment of the column.
	Comment *string

	// Reversible will allow the migrator to reverse the operation.
	Reversible *ColumnCommentOptions
}

type PrimaryKeyConstraintOptions struct {
	Table TableName

	Columns []string

	// IfNotExists Silently ignore if the primary key already exists, rather than raise an error.
	IfNotExists bool
}

func (p PrimaryKeyConstraintOptions) ConstraintType() string {
	return "primary_key"
}

type TableOptions struct {
	Table TableName

	// IfNotExists create the table if it doesn't exist.
	IfNotExists bool

	PrimaryKeys []string

	// Option is at the end of the table creation.
	Option string
}

type ColumnType = string

const (
	ColumnTypeString   ColumnType = "string"
	ColumnTypeText     ColumnType = "text"
	ColumnTypeInteger  ColumnType = "integer"
	ColumnTypeBigInt   ColumnType = "bigint"
	ColumnTypeFloat    ColumnType = "float"
	ColumnTypeDecimal  ColumnType = "decimal"
	ColumnTypeNumeric  ColumnType = "numeric"
	ColumnTypeDatetime ColumnType = "datetime"
	ColumnTypeTime     ColumnType = "time"
	ColumnTypeDate     ColumnType = "date"
	ColumnTypeBinary   ColumnType = "binary"
	ColumnTypeBlob     ColumnType = "blob"
	ColumnTypeBoolean  ColumnType = "boolean"
	ColumnTypeUUID     ColumnType = "uuid"
	ColumnTypeJSON     ColumnType = "json"

	// ColumnTypePrimaryKey is a special column type for primary keys.
	ColumnTypePrimaryKey ColumnType = "primary_key"

	// Postgres specific ----------------------------

	ColumnTypeSmallSerial ColumnType = "smallserial"
	ColumnTypeSerial      ColumnType = "serial"
	ColumnTypeBigSerial   ColumnType = "bigserial"
	ColumnTypeJSONB       ColumnType = "jsonb"
	ColumnTypeHstore      ColumnType = "hstore"

	// End of Postgres specific ----------------------------
)
