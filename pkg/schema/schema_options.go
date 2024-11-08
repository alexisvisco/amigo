package schema

import (
	"fmt"
	"strings"

	"github.com/alexisvisco/amigo/pkg/utils"
)

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

	// Validate Specify whether the constraint should be validated. Defaults to true.
	// Postgres only.
	Validate *bool
}

// ConstraintType returns the type of the constraint.
func (c CheckConstraintOptions) ConstraintType() string {
	return "check"
}

func (c CheckConstraintOptions) EventName() string {
	return "CheckConstraintEvent"
}

func (c CheckConstraintOptions) String() string {
	return fmt.Sprintf("-- add_check_constraint(table: %s, name: %s, expression: %s)", c.Table, c.ConstraintName,
		c.Expression)
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

type DropCheckConstraintOptions struct {
	Table          TableName
	ConstraintName string

	// ConstraintNameBuilder will build the name of the constraint. If nil, a default name will be used.
	// By default, is nil.
	ConstraintNameBuilder func(table TableName, constraintName string) string

	// IfExists add checks if the constraint exists before dropping it.
	IfExists bool

	// Reversible will allow the migrator to reverse the operation by creating the constraint.
	Reversible *CheckConstraintOptions
}

func (d DropCheckConstraintOptions) BuildConstraintName(table TableName, constraintName string) string {
	if d.ConstraintName != "" {
		return d.ConstraintName
	}

	if d.ConstraintNameBuilder == nil {
		return DefaultCheckConstraintNameBuilder(table, constraintName)
	}

	return d.ConstraintNameBuilder(table, constraintName)
}

func (d DropCheckConstraintOptions) EventName() string {
	return "DropCheckConstraintEvent"
}

func (d DropCheckConstraintOptions) String() string {
	return fmt.Sprintf("-- drop_check_constraint(table: %s, name: %s)", d.Table, d.ConstraintName)
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

	// Validate specifies whether the constraint should be validated. Defaults to true.
	// Postgres only.
	Validate *bool

	// Deferrable specifies whether the foreign key should be deferrable.
	// Could be DEFERRABLE, NOT DEFERRABLE, INITIALLY DEFERRED, INITIALLY IMMEDIATE or both.
	// Postgres only.
	Deferrable string
}

// ConstraintType returns the type of the constraint.
func (f AddForeignKeyConstraintOptions) ConstraintType() string {
	return "foreign_key"
}

func (f AddForeignKeyConstraintOptions) EventName() string {
	return "ForeignKeyEvent"
}

func (f AddForeignKeyConstraintOptions) String() string {
	return fmt.Sprintf("-- add_foreign_key_constraint(from: %s, to: %s, name: %s)", f.FromTable, f.ToTable,
		f.ForeignKeyName)
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

type DropForeignKeyConstraintOptions struct {
	FromTable      TableName
	ToTable        TableName
	ForeignKeyName string

	// ForeignKeyNameBuilder will build the name of the foreign key. If nil, a default name will be used.
	ForeignKeyNameBuilder func(fromTable TableName, toTable TableName) string

	// IfExists check if the foreign key exists before dropping it.
	IfExists bool

	// Reversible will allow the migrator to reverse the operation by creating the foreign key.
	Reversible *AddForeignKeyConstraintOptions
}

func (d DropForeignKeyConstraintOptions) BuildForeignKeyName(fromTable TableName, toTable TableName) string {
	if d.ForeignKeyName != "" {
		return d.ForeignKeyName
	}

	if d.ForeignKeyNameBuilder == nil {
		return DefaultForeignKeyNameBuilder(fromTable, toTable)
	}

	return d.ForeignKeyNameBuilder(fromTable, toTable)
}

func (d DropForeignKeyConstraintOptions) EventName() string {
	return "DropForeignKeyEvent"
}

func (d DropForeignKeyConstraintOptions) String() string {
	return fmt.Sprintf("-- drop_foreign_key_constraint(table: %s, name: %s)", d.FromTable, d.ForeignKeyName)
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

	// Order specifies the order of the index.
	Order string

	// OrderPerColumn specifies the order of the index per column.
	OrderPerColumn map[string]string

	// Predicate specifies the predicate for the index.
	// Postgres, SQLite only.
	Predicate string

	// Method specifies the index method. USING {method}.
	// Postgres only.
	Method string

	// Concurrent specifies if the index should be created concurrently.
	// Postgres only.
	Concurrent bool
}

func (i IndexOptions) EventName() string {
	return "IndexEvent"
}

func (i IndexOptions) String() string {
	return fmt.Sprintf("-- add_index(table: %s, name: %s, columns: %s)", i.Table, i.IndexName, i.Columns)
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
	Table     TableName
	Columns   []string
	IndexName string

	// IndexNameBuilder will build the name of the index. If nil, a default name will be used.
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

func (d DropIndexOptions) EventName() string {
	return "DropIndexEvent"
}

func (d DropIndexOptions) String() string {
	return fmt.Sprintf("-- drop_index(table: %s, name: %s)", d.Table, d.IndexName)
}

type ExtensionOptions struct {
	ExtensionName string

	// Schema is the schema where the extension will be created.
	Schema string

	// IfNotExists add IF NOT EXISTS to the query.
	IfNotExists bool

	Reversible *DropExtensionOptions
}

func (e ExtensionOptions) EventName() string {
	return "ExtensionEvent"
}

func (e ExtensionOptions) String() string {
	return fmt.Sprintf("-- add_extension(%s)", e.ExtensionName)
}

type DropExtensionOptions struct {
	ExtensionName string

	// IfExists add IF EXISTS to the query.
	IfExists bool

	// Reversible will allow the migrator to reverse the operation by creating the extension.
	Reversible *ExtensionOptions

	Cascade bool
}

func (e DropExtensionOptions) EventName() string {
	return "DropExtensionEvent"
}

func (e DropExtensionOptions) String() string {
	return fmt.Sprintf("-- drop_extension(%s)", e.ExtensionName)
}

type ColumnData interface {
	SetLimit(int)
	GetLimit() int
	GetPrecision() int
	SetNotNull(bool)
	SetPrimaryKey(bool)
	SetPrecision(int)
	GetType() ColumnType
	GetScale() int
	IsArray() bool
}

type ColumnOptions struct {
	Table      TableName
	ColumnName string

	// PrimaryKey specifies if the column is a primary key.
	// It will automatically add a PrimaryKeyConstraintOptions to the Constraints list.
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

	// Limit is a maximum column length. This is the number of characters for a varchar column.
	Limit int

	// IfNotExists specifies if the column already exists to not try to re-add it. This will avoid duplicate column errors.
	IfNotExists bool

	// Array specifies if the column is an array.
	Array bool

	// Comment is the Comment of the column.
	Comment string

	// Constraints is a list of constraints for the column.
	Constraints []ConstraintOption
}

func (c *ColumnOptions) EventName() string {
	return "ColumnEvent"
}

func (c *ColumnOptions) String() string {
	return fmt.Sprintf("-- add_column(table: %s, column: %s, type: %s)", c.Table, c.ColumnName, c.ColumnType)
}

func (c *ColumnOptions) SetLimit(limit int) {
	c.Limit = limit
}

func (c *ColumnOptions) SetNotNull(notNull bool) {
	c.NotNull = notNull
}

func (c *ColumnOptions) SetPrimaryKey(primaryKey bool) {
	c.PrimaryKey = primaryKey
}

func (c *ColumnOptions) SetPrecision(precision int) {
	c.Precision = precision
}

func (c *ColumnOptions) GetPrecision() int {
	return c.Precision
}

func (c *ColumnOptions) GetLimit() int {
	return c.Limit
}

func (c *ColumnOptions) GetType() ColumnType {
	return c.ColumnType
}

func (c *ColumnOptions) GetScale() int {
	return c.Scale
}

func (c *ColumnOptions) IsArray() bool {
	return c.Array
}

type DropColumnOptions struct {
	Table      TableName
	ColumnName string

	// IfExists add IF EXISTS to the query.
	IfExists bool

	// Reversible will allow the migrator to reverse the operation by creating the column.
	Reversible *ColumnOptions
}

func (d DropColumnOptions) EventName() string {
	return "DropColumnEvent"
}

func (d DropColumnOptions) String() string {
	return fmt.Sprintf("-- drop_column(table: %s, column: %s)", d.Table, d.ColumnName)
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

func (c ColumnCommentOptions) EventName() string {
	return "ColumnCommentEvent"
}

func (c ColumnCommentOptions) String() string {
	cmt := "NULL"
	if c.Comment != nil {
		cmt = fmt.Sprintf("%q", *c.Comment)
	}
	return fmt.Sprintf("-- comment_column(table: %s, column: %s, comment: %s)", c.Table, c.ColumnName, cmt)
}

type RenameColumnOptions struct {
	Table         TableName
	OldColumnName string
	NewColumnName string
}

func (r RenameColumnOptions) EventName() string {
	return "RenameColumnEvent"
}

func (r RenameColumnOptions) String() string {
	return fmt.Sprintf("-- rename_column(%s, old: %s, new: %s)", r.Table, r.OldColumnName, r.NewColumnName)
}

type RenameTableOptions struct {
	OldTable TableName
	NewTable TableName
}

func (r RenameTableOptions) EventName() string {
	return "RenameTableEvent"
}

func (r RenameTableOptions) String() string {
	return fmt.Sprintf("-- rename_table(old: %s, new: %s)", r.OldTable, r.NewTable)
}

type ChangeColumnTypeOptions struct {
	Table      TableName
	ColumnName string

	// ColumnType is the type of the column.
	// Could be a custom one but it's recommended to use the predefined ones for portability.
	ColumnType ColumnType

	// Limit is a maximum column length. This is the number of characters for a ColumnTypeString column and number of
	// bytes for ColumnTypeText, ColumnTypeBinary, ColumnTypeBlob, and ColumnTypeInteger Columns.
	Limit int

	// Scale is the scale of the column.
	// Mainly supported for Specifies the scale for the ColumnTypeDecimal, ColumnTypeNumeric
	// The scale is the number of digits to the right of the decimal point in a number.
	Scale int

	// Precision is the precision of the column.
	// Mainly supported for the ColumnTypeDecimal, ColumnTypeNumeric, ColumnTypeDatetime, and ColumnTypeTime
	// The precision is the number of significant digits in a number.
	Precision int

	// Array specifies if the column is an array.
	Array bool

	// Using is the USING clause for the change column type.
	// Postgres only.
	Using string

	// Reversible will allow the migrator to reverse the operation by creating the column.
	// Specify the old column type.
	Reversible *ChangeColumnTypeOptions
}

func (c *ChangeColumnTypeOptions) EventName() string {
	return "ChangeColumnTypeEvent"
}

func (c *ChangeColumnTypeOptions) String() string {
	return fmt.Sprintf("-- change_column_type(table: %s, column: %s, type: %s)", c.Table, c.ColumnName, c.ColumnType)
}

func (c *ChangeColumnTypeOptions) SetLimit(limit int) {
	c.Limit = limit
}

func (c *ChangeColumnTypeOptions) SetNotNull(_ bool) {
	// Do nothing
}

func (c *ChangeColumnTypeOptions) SetPrecision(precision int) {
	c.Precision = precision
}

func (c *ChangeColumnTypeOptions) GetPrecision() int {
	return c.Precision
}

func (c *ChangeColumnTypeOptions) SetPrimaryKey(_ bool) {
	// Do nothing
}

func (c *ChangeColumnTypeOptions) GetLimit() int {
	return c.Limit
}

func (c *ChangeColumnTypeOptions) GetType() ColumnType {
	return c.ColumnType
}

func (c *ChangeColumnTypeOptions) GetScale() int {
	return c.Scale
}

func (c *ChangeColumnTypeOptions) IsArray() bool {
	return c.Array
}

type ChangeColumnDefaultOptions struct {
	Table      TableName
	ColumnName string
	Value      string

	Reversible *ChangeColumnDefaultOptions
}

func (c *ChangeColumnDefaultOptions) EventName() string {
	return "ChangeColumnDefaultEvent"
}

func (c *ChangeColumnDefaultOptions) String() string {
	return fmt.Sprintf("-- change_column_default(table: %s, column: %s, value: %s)", c.Table, c.ColumnName, c.Value)
}

type TableCommentOptions struct {
	Table   TableName
	Comment *string

	Reversible *TableCommentOptions
}

func (t TableCommentOptions) EventName() string {
	return "TableCommentEvent"
}

func (t TableCommentOptions) String() string {
	cmt := "NULL"
	if t.Comment != nil {
		cmt = fmt.Sprintf("%q", *t.Comment)
	}
	return fmt.Sprintf("-- comment_table(table: %s, comment: %s)", t.Table, cmt)
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

func (p PrimaryKeyConstraintOptions) EventName() string {
	return "PrimaryKeyEvent"
}

func (p PrimaryKeyConstraintOptions) String() string {
	return fmt.Sprintf("-- add_primary_key_constraint(table: %s, columns: %s)", p.Table, strings.Join(p.Columns, ", "))
}

var DefaultPrimaryKeyNameBuilder = func(table TableName) string {
	return fmt.Sprintf("%s_pkey", table.Name())
}

type DropPrimaryKeyConstraintOptions struct {
	Table TableName

	PrimaryKeyName string

	PrimaryKeyNameBuilder func(table TableName) string

	// IfExists add IF EXISTS to the query.
	IfExists bool

	// Reversible will allow the migrator to reverse the operation by creating the primary key.
	Reversible *PrimaryKeyConstraintOptions
}

func (d DropPrimaryKeyConstraintOptions) BuildPrimaryKeyName(table TableName) string {
	if d.PrimaryKeyName != "" {
		return d.PrimaryKeyName
	}

	if d.PrimaryKeyNameBuilder == nil {
		return DefaultPrimaryKeyNameBuilder(table)
	}

	return d.PrimaryKeyNameBuilder(table)
}

func (d DropPrimaryKeyConstraintOptions) EventName() string {
	return "DropPrimaryKeyEvent"
}

func (d DropPrimaryKeyConstraintOptions) String() string {
	return fmt.Sprintf("-- drop_primary_key_constraint(table: %s, name: %s)", d.Table, d.PrimaryKeyName)
}

type TableDef interface {
	Columns() []ColumnOptions
	AfterTableCreate() []func()
}

type CreateEnumOptions struct {
	Name   string
	Values []string
	Schema string
}

func (c CreateEnumOptions) EventName() string {
	return "CreateEnumEvent"
}

func (c CreateEnumOptions) String() string {
	return fmt.Sprintf("-- create_enum(%s, %v)", c.Name, c.Values)
}

type DropEnumOptions struct {
	Name   string
	Schema string

	// IfExists add IF EXISTS to the query.
	IfExists bool

	// Reversible will allow the migrator to reverse the operation by creating the enum.
	Reversible *CreateEnumOptions
}

func (d DropEnumOptions) EventName() string {
	return "DropEnumEvent"
}

func (d DropEnumOptions) String() string {
	return fmt.Sprintf("-- drop_enum(%s)", d.Name)
}

type EnumUsage struct {
	Table  TableName
	Column string
}

type AddEnumValueOptions struct {
	Name   string
	Schema string

	Value string

	BeforeValue string
	AfterValue  string
}

func (a AddEnumValueOptions) EventName() string {
	return "AddEnumValueEvent"
}

func (a AddEnumValueOptions) String() string {
	return fmt.Sprintf("-- add_enum_value(%s, %s)", a.Name, a.Value)
}

type RenameEnumOptions struct {
	OldName string
	Schema  string

	NewName string
}

func (r RenameEnumOptions) EventName() string {
	return "RenameEnumEvent"
}

func (r RenameEnumOptions) String() string {
	return fmt.Sprintf("-- rename_enum(old: %s, new: %s)", r.OldName, r.NewName)
}

type RenameEnumValueOptions struct {
	Name   string
	Schema string

	OldValue string
	NewValue string
}

type TableOptions struct {
	Table TableName

	// IfNotExists create the table if it doesn't exist.
	IfNotExists bool

	// PrimaryKeys is a list of primary keys.
	PrimaryKeys []string

	// WithoutPrimaryKey specifies if the table should be created without a primary key.
	// By default, if you create a table without an "id" column and PrimaryKeys is empty it will not fail.
	// But if you want to explicitly create a table without a primary key, you can set this to true.
	WithoutPrimaryKey bool

	// Option is at the end of the table creation.
	Option string

	Comment *string

	// TableDefinition is the definition of the table. Usually a struct that implements TableDef will allow you to
	// define the columns and other options.
	TableDefinition TableDef
}

func (s TableOptions) EventName() string {
	return "CreateTableEvent"
}

func (s TableOptions) String() string {
	columns := utils.Map(s.TableDefinition.Columns(), func(c ColumnOptions) string {
		return fmt.Sprintf("%s", c.ColumnName)
	})
	return fmt.Sprintf("-- create_table(table: %s, {columns: %s}, {pk: %s})",
		s.Table,
		strings.Join(columns, ", "), strings.Join(s.PrimaryKeys, ", "))
}

type DropTableOptions struct {
	Table TableName

	// IfExists add IF EXISTS to the query.
	IfExists bool

	// Reversible will allow the migrator to reverse the operation by creating the table.
	Reversible *TableOptions
}

func (d DropTableOptions) EventName() string {
	return "DropTableEvent"
}

func (d DropTableOptions) String() string {
	return fmt.Sprintf("-- drop_table(table: %s)", d.Table)
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

	// ColumnTypeSmallSerial is an auto-incrementing integer column. Postgres only.
	ColumnTypeSmallSerial ColumnType = "smallserial"

	// ColumnTypeSerial is an auto-incrementing integer column. Postgres only.
	ColumnTypeSerial ColumnType = "serial"

	// ColumnTypeBigSerial is an auto-incrementing integer column. Postgres only.
	ColumnTypeBigSerial ColumnType = "bigserial"

	// ColumnTypeJSONB is a binary JSON column. Postgres only.
	ColumnTypeJSONB ColumnType = "jsonb"

	// ColumnTypeHstore is a key-value store column. Postgres only.
	ColumnTypeHstore ColumnType = "hstore"
)
