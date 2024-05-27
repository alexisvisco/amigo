package pg

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"strings"
)

// CreateEnum create a new enum type
// Is auto reversible
// Be careful because you can't remove an enum value easily.
//
// Example:
//
//	schema.CreateEnum("my_enum", []string{"a", "b", "c"})
//
// Generates:
//
//	CREATE TYPE my_enum AS ENUM ('a', 'b', 'c');
func (s *Schema) CreateEnum(name string, values []string, opts ...schema.CreateEnumOptions) {
	options := schema.CreateEnumOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Name = name
	options.Values = values

	if s.Context.MigrationDirection == types.MigrationDirectionDown {
		s.rollbackMode().DropEnum(name, schema.DropEnumOptions{
			Schema:   options.Schema,
			IfExists: true,
		})
		return
	}

	q := `CREATE TYPE {enum_name} AS ENUM ({values})`

	replacer := utils.Replacer{
		"enum_name": utils.StrFunc(formatEnumName(name, options.Schema)),
		"values": func() string {
			var valuesStr []string
			for _, v := range values {
				valuesStr = append(valuesStr, QuoteValue(v))
			}
			return strings.Join(valuesStr, ", ")
		},
	}

	_, err := s.DB.ExecContext(s.Context.Context, replacer.Replace(q))
	if err != nil {
		s.Context.RaiseError(fmt.Errorf("error while creating enum: %w", err))
		return
	}

	s.Context.AddEnumCreated(schema.CreateEnumOptions{Name: name, Values: values})
}

// AddEnumValue add a new value to an existing enum type
// This operation is not reversible. https://www.postgresql.org/message-id/21012.1459434338%40sss.pgh.pa.us
//
// WARNINGS: This operation is not running in a transaction. So it can't be rolled back.
//
// Example:
//
//	schema.AddEnumValue("my_enum", "d")
//
// Generates:
//
//	ALTER TYPE my_enum ADD VALUE 'd';
func (s *Schema) AddEnumValue(name string, value string, opts ...schema.AddEnumValueOptions) {
	options := schema.AddEnumValueOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Name = name
	options.Value = value

	if s.Context.MigrationDirection == types.MigrationDirectionDown {
		logger.Warn(events.MessageEvent{
			Message: fmt.Sprintf("it is not possible to reverse adding enum value %s to %s", value, name),
		})

		return
	}

	q := `ALTER TYPE {enum_name} ADD VALUE {value} {before_value} {after_value}`

	replacer := utils.Replacer{
		"enum_name": utils.StrFunc(formatEnumName(name, options.Schema)),
		"value":     utils.StrFunc(QuoteValue(value)),

		"before_value": utils.StrFuncPredicate(options.BeforeValue != "",
			fmt.Sprintf("BEFORE %s", QuoteValue(options.BeforeValue))),
		"after_value": utils.StrFuncPredicate(options.AfterValue != "",
			fmt.Sprintf("AFTER %s", QuoteValue(options.AfterValue))),
	}

	_, err := s.NotInTx.ExecContext(s.Context.Context, replacer.Replace(q))
	if err != nil {
		s.Context.RaiseError(fmt.Errorf("error while adding enum value: %w", err))
		return
	}

	s.Context.AddEnumValueCreated(schema.AddEnumValueOptions{Name: name, Value: value})
}

// DropEnum drop an enum type
// Must precise options.Reversible to be able to reverse the operation
//
// Example:
//
//	schema.DropEnum("my_enum")
//
// Generates:
//
//	DROP TYPE my_enum;
func (s *Schema) DropEnum(name string, opts ...schema.DropEnumOptions) {
	options := schema.DropEnumOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Name = name

	if s.Context.MigrationDirection == types.MigrationDirectionUp {
		if options.Reversible != nil {
			s.rollbackMode().CreateEnum(name, options.Reversible.Values)
		} else {
			logger.Warn(events.MessageEvent{
				Message: fmt.Sprintf("unable to reverse dropping enum %s", name),
			})
		}
		return
	}

	q := `DROP TYPE {if_exists} {enum_name}`

	replacer := utils.Replacer{
		"enum_name": utils.StrFunc(formatEnumName(name, options.Schema)),
		"if_exists": utils.StrFuncPredicate(options.IfExists, "IF EXISTS"),
	}

	_, err := s.DB.ExecContext(s.Context.Context, replacer.Replace(q))
	if err != nil {
		s.Context.RaiseError(fmt.Errorf("error while dropping enum: %w", err))
		return
	}

	s.Context.AddEnumDropped(schema.DropEnumOptions{Name: name})

}

// FindEnumUsage find all tables and columns that use the enum type
// schemaName is optional, if provided, it will restrict the search to the specified schema
func (s *Schema) FindEnumUsage(name string, schemaName *string) []schema.EnumUsage {
	var usages []schema.EnumUsage

	q := `SELECT 
    n.nspname AS schema_name,
    c.relname AS table_name,
    a.attname AS column_name
FROM 
    pg_attribute a
    JOIN pg_class c ON a.attrelid = c.oid
    JOIN pg_type t ON a.atttypid = t.oid
    JOIN pg_namespace n ON c.relnamespace = n.oid
WHERE 
    t.typcategory = 'E'  -- Enum type category
    AND t.typname = $1  -- Enum type name
    
    -- Optional: to restrict to a specific schema
    {filter_schema}
ORDER BY 
    schema_name, table_name, column_name`

	values := []interface{}{name}

	replacer := utils.Replacer{
		"filter_schema": utils.StrFuncPredicate(schemaName != nil, "AND n.nspname = $2"),
	}

	if schemaName != nil {
		values = append(values, *schemaName)
	}

	rows, err := s.DB.QueryContext(s.Context.Context, replacer.Replace(q), values...)
	if err != nil {
		s.Context.RaiseError(fmt.Errorf("error while finding enum usage: %w", err))
		return nil
	}

	defer rows.Close()

	for rows.Next() {
		var usage schema.EnumUsage
		var sc, table string
		err := rows.Scan(&sc, &table, &usage.Column)
		if err != nil {
			s.Context.RaiseError(fmt.Errorf("error while scanning enum usage: %w", err))
			return nil
		}
		usage.Table = schema.TableName(fmt.Sprintf("%s.%s", sc, table))
		usages = append(usages, usage)
	}

	return usages
}

func (s *Schema) ListEnumValues(name string, schemaName *string) []string {
	var values []string

	q := `SELECT enumlabel
FROM pg_enum
	JOIN pg_type ON pg_enum.enumtypid = pg_type.oid
	JOIN pg_namespace ON pg_type.typnamespace = pg_namespace.oid
WHERE typname = $1
	{filter_schema}
ORDER BY enumsortorder`

	args := []any{name}

	replacer := utils.Replacer{
		"filter_schema": utils.StrFuncPredicate(schemaName != nil, "AND nspname = $2"),
	}

	if schemaName != nil {
		args = append(args, *schemaName)
	}

	rows, err := s.DB.QueryContext(s.Context.Context, replacer.Replace(q), args...)
	if err != nil {
		s.Context.RaiseError(fmt.Errorf("error while listing enum values: %w", err))
		return nil
	}

	defer rows.Close()

	for rows.Next() {
		var value string
		err := rows.Scan(&value)
		if err != nil {
			s.Context.RaiseError(fmt.Errorf("error while scanning enum value: %w", err))
			return nil
		}
		values = append(values, value)
	}

	return values
}

// RenameEnum rename an enum type
// Example:
//
//	schema.RenameEnum("old_enum", "new_enum")
//
// Generates:
//
//	ALTER TYPE old_enum RENAME TO new_enum
func (s *Schema) RenameEnum(oldName, newName string, opts ...schema.RenameEnumOptions) {
	options := schema.RenameEnumOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.OldName = oldName
	options.NewName = newName

	if s.Context.MigrationDirection == types.MigrationDirectionDown {
		s.rollbackMode().RenameEnum(newName, oldName)
		return
	}

	q := `ALTER TYPE {old_enum_name} RENAME TO {new_enum_name}`

	replacer := utils.Replacer{
		"old_enum_name": utils.StrFunc(formatEnumName(oldName, options.Schema)),
		"new_enum_name": utils.StrFunc(QuoteIdent(newName)),
	}

	_, err := s.DB.ExecContext(s.Context.Context, replacer.Replace(q))
	if err != nil {
		s.Context.RaiseError(fmt.Errorf("error while renaming enum: %w", err))
		return
	}

	s.Context.AddRenameEnum(schema.RenameEnumOptions{OldName: oldName, NewName: newName})
}

// RenameEnumValue rename an enum value
// Is auto reversible
//
// Example:
//
//	schema.RenameEnumValue("my_enum", "old_value", "new_value")
//
// Generates:
//
//	ALTER TYPE my_enum RENAME VALUE 'old_value' TO 'new_value'
func (s *Schema) RenameEnumValue(name, oldName, newName string, opts ...schema.RenameEnumValueOptions) {
	options := schema.RenameEnumValueOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Name = name
	options.OldValue = oldName
	options.NewValue = newName

	if s.Context.MigrationDirection == types.MigrationDirectionDown {
		s.rollbackMode().RenameEnumValue(name, newName, oldName, options)
		return
	}

	q := `ALTER TYPE {enum_name} RENAME VALUE {old_value} TO {new_value}`

	replacer := utils.Replacer{
		"enum_name": utils.StrFunc(formatEnumName(name, options.Schema)),
		"old_value": utils.StrFunc(QuoteValue(oldName)),
		"new_value": utils.StrFunc(QuoteValue(newName)),
	}

	_, err := s.DB.ExecContext(s.Context.Context, replacer.Replace(q))
	if err != nil {
		s.Context.RaiseError(fmt.Errorf("error while renaming enum value: %w", err))
		return
	}

	s.Context.AddRenameEnumValue(options)
}

func formatEnumName(name string, schema string) string {
	if schema != "" {
		return fmt.Sprintf("%s.%s", QuoteIdent(schema), QuoteIdent(name))
	}
	return QuoteIdent(name)
}
