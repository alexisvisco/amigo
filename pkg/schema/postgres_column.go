package schema

import (
	"fmt"
	"strings"
)

// AddColumn adds a new column to the Table.
//
// Example:
//
//	p.AddColumn("users", "picture", ColumnTypeBinary)
//
// Generates:
//
//	ALTER TABLE "users" ADD "picture" BYTEA
//
// Adding a column with a limit, default value and null constraint:
//
//	p.AddColumn("articles", "status", ColumnTypeString, ColumnOptions{Limit: 20, Default: "draft", NotNull: false})
//
// Generates:
//
//	ALTER TABLE "articles" ADD "status" VARCHAR(20) DEFAULT 'draft' NOT NULL
//
// Adding a column with precision and scale:
//
//	p.AddColumn("answers", "bill_gates_money", ColumnTypeDecimal, ColumnOptions{Precision: 15, Scale: 2})
//
// Generates:
//
//	ALTER TABLE "answers" ADD "bill_gates_money" DECIMAL(15,2)
//
// Adding a column with an array type:
//
//	p.AddColumn("users", "skills", ColumnTypeText, ColumnOptions{Array: true})
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
//	p.AddColumn("shapes", "triangle", "polygon", ColumnOptions{IfNotExists: true})
//
// Generates:
//
//	ALTER TABLE "shapes" ADD "triangle" IF NOT EXISTS POLYGON
func (p *Postgres) AddColumn(tableName TableName, columnName string, columnType ColumnType, options ColumnOptions) {
	options.Table = tableName
	options.ColumnName = columnName
	options.ColumnType = p.toType(columnType, &options)
	if options.PrimaryKey {
		options.NotNull = true
		options.Constraints = append(options.Constraints, PrimaryKeyConstraintOptions{})
	}

	if p.Context.migrationType == MigrationTypeDown {
		// todo: implement down migration
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD %s", options.Table, p.column(options))

	_, err := p.db.ExecContext(p.Context.Context, query)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding column: %w", err))
		return
	}

	p.Context.addColumnCreated(options)

	if options.Comment != "" {
		p.AddColumnComment(options.Table, options.ColumnName, &options.Comment, ColumnCommentOptions{})
	}
}

func (p *Postgres) column(options ColumnOptions) string {
	sql := `{if_not_exists} "{column_name}" {column_type} {default} {nullable} {constraints}`

	replacer := replacer{
		"column_name": strfunc(options.ColumnName),
		"column_type": func() string {
			strBuilder := strings.Builder{}
			strBuilder.WriteString(options.ColumnType)
			if options.Limit > 0 {
				if strings.ToLower(options.ColumnType) == "varchar" {
					strBuilder.WriteString(fmt.Sprintf("(%d)", options.Limit))
				}
			} else {

				// would add precision and scale for decimal and numeric types
				// example : DECIMAL(15,2) where 15 is precision and 2 is scale
				// The scale cannot be set without setting the precision

				precisionAndScale := make([]string, 0, 2)
				if options.Precision > 0 {
					precisionAndScale = append(precisionAndScale, fmt.Sprintf("%d", options.Precision))
				}

				if options.Scale > 0 && options.Precision == 0 {
					p.Context.RaiseError(fmt.Errorf("scale cannot be set without setting the precision"))
					return ""
				}

				if options.Scale > 0 {
					precisionAndScale = append(precisionAndScale, fmt.Sprintf("%d", options.Scale))
				}

				if len(precisionAndScale) > 0 {
					strBuilder.WriteString(fmt.Sprintf("(%s)", strings.Join(precisionAndScale, ", ")))
				}
			}

			if options.Array {
				strBuilder.WriteString("[]")
			}

			return strBuilder.String()
		},
		"if_not_exists": func() string {
			if options.IfNotExists {
				return "IF NOT EXISTS"
			}
			return ""
		},
		"default": func() string {
			if options.Default != "" {
				return fmt.Sprintf("DEFAULT '%s'", options.Default)
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

	return replacer.replace(sql)
}

// AddColumnComment adds a comment to the column.
//
// Example:
//
//	p.AddColumnComment("users", "name", schema.Ptr("The name of the user"))
//
// Generates:
//
//	COMMENT ON COLUMN "users"."name" IS 'The name of the user'
//
// To set a null comment:
//
//	p.AddColumnComment("users", "name", nil)
//
// Generates:
//
//	COMMENT ON COLUMN "users"."name" IS NULL
//
// To be able to reverse the operation you must provide the Reversible parameter
func (p *Postgres) AddColumnComment(tableName TableName, columnName string, comment *string, options ColumnCommentOptions) {
	options.Table = tableName
	options.ColumnName = columnName
	options.Comment = comment

	if p.Context.migrationType == MigrationTypeDown && options.Reversible != nil {
		p.executeAddColumnComment(*options.Reversible)
		return
	}

	p.executeAddColumnComment(options)
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
func (p *Postgres) RenameColumn(tableName TableName, oldColumnName, newColumnName string) {
	if p.Context.migrationType == MigrationTypeDown {
		temp := oldColumnName
		oldColumnName = newColumnName
		newColumnName = temp
	}

	query := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", tableName, oldColumnName, newColumnName)

	_, err := p.db.ExecContext(p.Context.Context, query)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while renaming column: %w", err))
		return
	}
}

func (p *Postgres) executeAddColumnComment(options ColumnCommentOptions) {
	sql := `COMMENT ON COLUMN {table_name}.{column_name} IS {comment}`

	replacer := replacer{
		"table_name":  strfunc(options.Table.String()),
		"column_name": strfunc(options.ColumnName),
		"comment": func() string {
			if options.Comment == nil {
				return "NULL"
			}
			return fmt.Sprintf("'%s'", *options.Comment)
		},
	}

	_, err := p.db.ExecContext(p.Context.Context, replacer.replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding column comment: %w", err))
		return
	}
}

func (p *Postgres) toType(c ColumnType, co *ColumnOptions) string {
	p.modifyColumnOptionFromType(c, co)

	serialFunc := func() string {
		if co.Limit > 0 {
			if co.Limit <= 2 {
				return "SMALLSERIAL"
			} else if co.Limit <= 4 {
				return "SERIAL"
			} else {
				return "BIGSERIAL"
			}
		}

		return "SERIAL"
	}

	switch c {
	case ColumnTypeString:
		return "VARCHAR"
	case ColumnTypeText:
		return "TEXT"
	case ColumnTypeInteger:
		if co.Limit > 0 {
			if co.Limit <= 2 {
				return "SMALLINT"
			} else if co.Limit <= 4 {
				return "INTEGER"
			} else {
				return "BIGINT"
			}
		}
		return "INTEGER"
	case ColumnTypeSerial:
		return serialFunc()
	case ColumnTypeBigInt:
		return "BIGINT"
	case ColumnTypeFloat:
		return "FLOAT"
	case ColumnTypeDecimal, ColumnTypeNumeric:
		return "DECIMAL"
	case ColumnTypeDatetime:
		return "TIMESTAMP"
	case ColumnTypeTime:
		return "TIME"
	case ColumnTypeDate:
		return "DATE"
	case ColumnTypeBinary, ColumnTypeBlob:
		return "BYTEA"
	case ColumnTypeBoolean:
		return "BOOLEAN"
	case ColumnTypeUUID:
		return "UUID"
	case ColumnTypeJSON:
		return "JSON"
	case ColumnTypePrimaryKey:
		return serialFunc()
	case ColumnTypeJSONB:
		return "JSONB"
	case ColumnTypeHstore:
		return "HSTORE"

	default:
		return c
	}
}

func (p *Postgres) modifyColumnOptionFromType(c ColumnType, co *ColumnOptions) {
	switch c {
	case ColumnTypeBigSerial:
		co.Limit = 8
	case ColumnTypeSmallSerial:
		co.Limit = 2
	case ColumnTypePrimaryKey:
		co.NotNull = true
		co.PrimaryKey = true
	case ColumnTypeDatetime:
		if co.Precision == 0 {
			co.Precision = 6
		}
	}
}
