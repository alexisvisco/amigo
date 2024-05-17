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

// AddColumn adds a new column to the Table.
//
// Example:
//
//	p.AddColumn("users", "picture", schema.ColumnTypeBinary)
//
// Generates:
//
//	ALTER TABLE "users" ADD "picture" BYTEA
//
// Adding a column with a limit, default value and null constraint:
//
//	p.AddColumn("articles", "status", schema.ColumnTypeString, schema.ColumnOptions{Limit: 20, Default: "draft", NotNull: false})
//
// Generates:
//
//	ALTER TABLE "articles" ADD "status" VARCHAR(20) DEFAULT 'draft' NOT NULL
//
// Adding a column with precision and scale:
//
//	p.AddColumn("answers", "bill_gates_money", schema.ColumnTypeDecimal, schema.ColumnOptions{Precision: 15, Scale: 2})
//
// Generates:
//
//	ALTER TABLE "answers" ADD "bill_gates_money" DECIMAL(15,2)
//
// Adding a column with an array type:
//
//	p.AddColumn("users", "skills", schema.ColumnTypeText, schema.ColumnOptions{Array: true})
//
// Generates:
//
//	ALTER TABLE "users" ADD "skills" TEXT[]
//
// Adding a column with a custom type:
//
//	p.AddColumn("shapes", "triangle", "polygon")
//
// Generates:
//
//	ALTER TABLE "shapes" ADD "triangle" POLYGON
//
// Adding a column if it does not exist:
//
//	p.AddColumn("shapes", "triangle", "polygon", schema.ColumnOptions{IfNotExists: true})
//
// Generates:
//
//	ALTER TABLE "shapes" ADD "triangle" IF NOT EXISTS POLYGON
func (p *Schema) AddColumn(tableName schema.TableName, columnName string, columnType schema.ColumnType, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Table = tableName
	options.ColumnName = columnName
	options.ColumnType = p.toType(columnType, &options)
	if options.PrimaryKey {
		options.NotNull = true
		options.Constraints = append(options.Constraints, schema.PrimaryKeyConstraintOptions{})
	}

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		p.rollbackMode().DropColumn(tableName, columnName, schema.DropColumnOptions{IfExists: true})
		return
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD %s", options.Table, p.column(options))

	_, err := p.DB.ExecContext(p.Context.Context, query)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding column: %w", err))
		return
	}

	p.Context.AddColumnCreated(options)

	if options.Comment != "" {
		p.AddColumnComment(options.Table, options.ColumnName, &options.Comment, schema.ColumnCommentOptions{})
	}
}

func (p *Schema) column(options schema.ColumnOptions) string {
	sql := `{if_not_exists} "{column_name}" {column_type} {default} {nullable} {constraints}`

	replacer := utils.Replacer{
		"column_name": utils.StrFunc(options.ColumnName),
		"column_type": p.columnType(&options),
		"if_not_exists": func() string {
			if options.IfNotExists {
				return "IF NOT EXISTS"
			}
			return ""
		},
		"default": func() string {
			if options.Default != "" {
				return fmt.Sprintf("DEFAULT %s", options.Default)
			}
			return ""
		},
		"nullable": func() string {
			if options.NotNull {
				return "NOT NULL"
			}
			return ""
		},
		"constraints": func() string {
			var constraints []string

			for _, constraint := range options.Constraints {
				constraints = append(constraints, p.applyConstraint(constraint))
			}

			return strings.Join(constraints, " ")
		},
	}

	return replacer.Replace(sql)
}

func (p *Schema) columnType(options schema.ColumnData) func() string {
	return func() string {
		strBuilder := strings.Builder{}
		strBuilder.WriteString(options.GetType())
		if options.GetLimit() > 0 {
			if strings.ToLower(options.GetType()) == "varchar" {
				strBuilder.WriteString(fmt.Sprintf("(%d)", options.GetLimit()))
			}
		} else {

			// would add precision and scale for decimal and numeric types
			// example : DECIMAL(15,2) where 15 is precision and 2 is scale
			// The scale cannot be set without setting the precision

			precisionAndScale := make([]string, 0, 2)
			if options.GetPrecision() > 0 {
				precisionAndScale = append(precisionAndScale, fmt.Sprintf("%d", options.GetPrecision()))
			}

			if options.GetScale() > 0 && options.GetPrecision() == 0 {
				p.Context.RaiseError(fmt.Errorf("scale cannot be set without setting the precision"))
				return ""
			}

			if options.GetScale() > 0 {
				precisionAndScale = append(precisionAndScale, fmt.Sprintf("%d", options.GetScale()))
			}

			if len(precisionAndScale) > 0 {
				strBuilder.WriteString(fmt.Sprintf("(%s)", strings.Join(precisionAndScale, ", ")))
			}
		}

		if options.IsArray() {
			strBuilder.WriteString("[]")
		}

		return strBuilder.String()
	}
}

// AddTimestamps adds the timestamps columns created_at and updated_at to the table.
// It's a shortcut for adding two columns with the current timestamp as default value.
//
// Example:
//
//	p.AddTimestamps("users")
//
// Generates:
//
//	ALTER TABLE "users" ADD "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
//	ALTER TABLE "users" ADD "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
func (p *Schema) AddTimestamps(tableName schema.TableName) {
	p.AddColumn(tableName, "created_at", schema.ColumnTypeDatetime,
		schema.ColumnOptions{NotNull: true, Default: "now()"})
	p.AddColumn(tableName, "updated_at", schema.ColumnTypeDatetime,
		schema.ColumnOptions{NotNull: true, Default: "now()"})
}

// AddColumnComment adds a comment to the column.
//
// Example:
//
//	p.AddColumnComment("users", "name", schema.utils.Ptr("The name of the User"))
//
// Generates:
//
//	COMMENT ON COLUMN "users"."name" IS 'The name of the User'
//
// To set a null comment:
//
//	p.AddColumnComment("users", "name", nil)
//
// Generates:
//
//	COMMENT ON COLUMN "users"."name" IS NULL
//
// To be able to rollbackMode the operation you must provide the Reversible parameter
func (p *Schema) AddColumnComment(tableName schema.TableName, columnName string, comment *string, opts ...schema.ColumnCommentOptions) {
	options := schema.ColumnCommentOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Table = tableName
	options.ColumnName = columnName
	options.Comment = comment

	if p.Context.MigrationDirection == types.MigrationDirectionDown && options.Reversible != nil {
		p.rollbackMode().AddColumnComment(tableName, columnName, comment, schema.ColumnCommentOptions{})
		return
	}

	sql := `COMMENT ON COLUMN {table_name}.{column_name} IS {comment}`

	replacer := utils.Replacer{
		"table_name":  utils.StrFunc(options.Table.String()),
		"column_name": utils.StrFunc(options.ColumnName),
		"comment": func() string {
			if options.Comment == nil {
				return "NULL"
			}
			return fmt.Sprintf("'%s'", *options.Comment)
		},
	}

	_, err := p.DB.ExecContext(p.Context.Context, replacer.Replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding column comment: %w", err))
		return
	}

	p.Context.AddColumnComment(options)
}

// RenameColumn renames a column in the table.
// The column is renamed from oldColumnName to newColumnName.
//
// Example:
//
//	p.RenameColumn("users", "name", "full_name")
//
// Generates:
//
//	ALTER TABLE "users" RENAME COLUMN "name" TO "full_name"
func (p *Schema) RenameColumn(tableName schema.TableName, oldColumnName, newColumnName string) {
	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		temp := oldColumnName
		oldColumnName = newColumnName
		newColumnName = temp
	}

	query := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", tableName, oldColumnName, newColumnName)

	_, err := p.DB.ExecContext(p.Context.Context, query)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while renaming column: %w", err))
		return
	}

	p.Context.AddRenameColumn(schema.RenameColumnOptions{
		Table:         tableName,
		OldColumnName: oldColumnName,
		NewColumnName: newColumnName,
	})
}

// DropColumn drops a column from the table.
//
// Example:
//
//	p.DropColumn("users", "name")
//
// Generates:
//
//	ALTER TABLE "users" DROP COLUMN "name"
//
// Dropping a column if it exists:
//
//	p.DropColumn("users", "name", schema.DropColumnOptions{IfExists: true})
//
// Generates:
//
//	ALTER TABLE "users" DROP COLUMN IF EXISTS "name"
//
// To be able to reverse the operation you must provide the Reversible parameter:
//
//	p.DropColumn("users", "name", schema.DropColumnOptions{Reversible: &schema.ReversibleColumn{ColumnType: "VARCHAR(255)"}})
//
// Generates:
//
//	ALTER TABLE "users" ADD "name" VARCHAR(255)
func (p *Schema) DropColumn(tableName schema.TableName, columnName string, opt ...schema.DropColumnOptions) {
	options := schema.DropColumnOptions{}
	if len(opt) > 0 {
		options = opt[0]
	}

	options.Table = tableName
	options.ColumnName = columnName

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		if options.Reversible != nil {
			p.rollbackMode().AddColumn(tableName, columnName, options.Reversible.ColumnType, *options.Reversible)
		} else {
			logger.Warn(events.MessageEvent{
				Message: fmt.Sprintf("unable to recreate the column %s.%s", tableName, columnName),
			})
		}
		return
	}

	query := "ALTER TABLE {table_name} DROP COLUMN {if_exists} {column_name}"
	replacer := utils.Replacer{
		"table_name":  utils.StrFunc(options.Table.String()),
		"column_name": utils.StrFunc(options.ColumnName),
		"if_exists":   utils.StrFuncPredicate(options.IfExists, "IF EXISTS"),
	}

	_, err := p.DB.ExecContext(p.Context.Context, replacer.Replace(query))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while dropping column: %w", err))
		return
	}

	p.Context.AddColumnDropped(options)
}

// ChangeColumnType changes the column type and options of a column.
//
// Example:
//
//	p.ChangeColumnType("users", "name", schema.ColumnTypeString, schema.ColumnOptions{Limit: 20})
//
// Generates:
//
//	ALTER TABLE "users" ALTER COLUMN "name" TYPE VARCHAR(20)
//
// Changing a column type with a custom cast:
// For a table like:
//
//	CREATE TABLE users (invoice_no TEXT);
//
//	p.ChangeColumnType("users", "invoice_no", schema.ColumnTypeInteger, schema.ChangeColumnTypeOptions{Using: "invoice_no::integer"})
//
// Generates:
//
//	ALTER TABLE "users" ALTER COLUMN "invoice_no" TYPE INTEGER USING invoice_no::integer
func (p *Schema) ChangeColumnType(tableName schema.TableName, columnName string, columnType schema.ColumnType, opts ...schema.ChangeColumnTypeOptions) {
	options := schema.ChangeColumnTypeOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Table = tableName
	options.ColumnName = columnName
	options.ColumnType = p.toType(columnType, &options)

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		if options.Reversible != nil {
			p.rollbackMode().ChangeColumnType(tableName, columnName, columnType, *options.Reversible)
		} else {
			logger.Warn(events.MessageEvent{
				Message: fmt.Sprintf("unable to recreate the column %s.%s", tableName, columnName),
			})
		}
		return
	}

	query := "ALTER TABLE {table_name} ALTER COLUMN {column_name} TYPE {column_type} {using}"
	replacer := utils.Replacer{
		"table_name":  utils.StrFunc(options.Table.String()),
		"column_name": utils.StrFunc(options.ColumnName),
		"column_type": p.columnType(&options),
		"using":       utils.StrFuncPredicate(options.Using != "", fmt.Sprintf("USING %s", options.Using)),
	}

	_, err := p.DB.ExecContext(p.Context.Context, replacer.Replace(query))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while changing column: %w", err))
		return
	}

	p.Context.AddChangeColumnType(options)
}

func (p *Schema) toType(c schema.ColumnType, co schema.ColumnData) string {
	p.modifyColumnOptionFromType(c, co)

	serialFunc := func() string {
		if co.GetLimit() > 0 {
			if co.GetLimit() <= 2 {
				return "SMALLSERIAL"
			} else if co.GetLimit() <= 4 {
				return "SERIAL"
			} else {
				return "BIGSERIAL"
			}
		}

		return "SERIAL"
	}

	switch c {
	case schema.ColumnTypeString:
		return "VARCHAR"
	case schema.ColumnTypeText:
		return "TEXT"
	case schema.ColumnTypeInteger:
		if co.GetLimit() > 0 {
			if co.GetLimit() <= 2 {
				return "SMALLINT"
			} else if co.GetLimit() <= 4 {
				return "INTEGER"
			} else {
				return "BIGINT"
			}
		}
		return "INTEGER"
	case schema.ColumnTypeSerial:
		return serialFunc()
	case schema.ColumnTypeBigInt:
		return "BIGINT"
	case schema.ColumnTypeFloat:
		return "FLOAT"
	case schema.ColumnTypeDecimal, schema.ColumnTypeNumeric:
		return "DECIMAL"
	case schema.ColumnTypeDatetime:
		return "TIMESTAMP"
	case schema.ColumnTypeTime:
		return "TIME"
	case schema.ColumnTypeDate:
		return "DATE"
	case schema.ColumnTypeBinary, schema.ColumnTypeBlob:
		return "BYTEA"
	case schema.ColumnTypeBoolean:
		return "BOOLEAN"
	case schema.ColumnTypeUUID:
		return "UUID"
	case schema.ColumnTypeJSON:
		return "json"
	case schema.ColumnTypePrimaryKey:
		return serialFunc()
	case schema.ColumnTypeJSONB:
		return "JSONB"
	case schema.ColumnTypeHstore:
		return "HSTORE"

	default:
		return c
	}
}

func (p *Schema) modifyColumnOptionFromType(c schema.ColumnType, co schema.ColumnData) {
	switch c {
	case schema.ColumnTypeBigSerial:
		co.SetLimit(8)
	case schema.ColumnTypeSmallSerial:
		co.SetLimit(8)
	case schema.ColumnTypePrimaryKey:
		co.SetNotNull(true)
		co.SetPrimaryKey(true)
	case schema.ColumnTypeDatetime:
		if co.GetPrecision() == 0 {
			co.SetPrecision(6)
		}
	}
}
