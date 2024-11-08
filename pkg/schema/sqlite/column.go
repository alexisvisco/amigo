package sqlite

import (
	"fmt"

	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/base"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
)

// AddColumn adds a new column to the Table.
// - You cannot add a primary key column to a table.
//
// Example:
//
//	p.Column("users", "picture", schema.ColumnTypeBinary)
//
// Generates:
//
//	ALTER TABLE "users" ADD "picture" BLOB
//
// Adding a column with a limit, default value and null constraint:
//
//	p.Column("articles", "status", schema.ColumnTypeString, schema.ColumnOptions{Limit: 20, Default: "draft", NotNull: false})
//
// Generates:
//
//	ALTER TABLE "articles" ADD "status" VARCHAR(20) DEFAULT 'draft' NOT NULL
//
// Adding a column with precision and scale:
//
//	p.Column("answers", "bill_gates_money", schema.ColumnTypeDecimal, schema.ColumnOptions{Precision: 15, Scale: 2})
//
// Generates:
//
//	ALTER TABLE "answers" ADD "bill_gates_money" NUMERIC(15,2)
//
// Adding a column if it does not exist:
//
//	p.Column("shapes", "triangle", "polygon", schema.ColumnOptions{IfNotExists: true})
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

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		//p.rollbackMode().DropColumn(tableName, columnName, schema.DropColumnOptions{IfExists: true})
		return
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD %s", options.Table, p.column(options))

	_, err := p.TX.ExecContext(p.Context.Context, query)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding column: %w", err))
		return
	}

	p.Context.AddColumnCreated(options)

	if options.Comment != "" {
		//p.AddColumnComment(options.Table, options.ColumnName, &options.Comment, schema.ColumnCommentOptions{})
	}
}

func (p *Schema) column(options schema.ColumnOptions) string {
	sql := `{if_not_exists} "{column_name}" {column_type} {primary_key} {default} {nullable}`

	replacer := utils.Replacer{
		"column_name": utils.StrFunc(options.ColumnName),
		"column_type": base.ColumnType(p.Context, &options),
		"if_not_exists": func() string {
			if options.IfNotExists {
				return "IF NOT EXISTS"
			}
			return ""
		},
		"primary_key": func() string {
			if options.PrimaryKey {
				s := ""
				if options.ColumnType == schema.ColumnTypePrimaryKey {
					s += "INTEGER"
				} else {
					s += options.ColumnType
				}
				s = "PRIMARY KEY"
				if options.ColumnType == schema.ColumnTypeSerial || options.ColumnType == schema.ColumnTypePrimaryKey {
					s += " AUTOINCREMENT"
				}

				return s
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
		//"constraints": func() string {
		//	var constraints []string
		//
		//	for _, constraint := range options.Constraints {
		//		constraints = append(constraints, p.applyConstraint(constraint))
		//	}
		//
		//	return strings.Join(constraints, " ")
		//},
	}

	return replacer.Replace(sql)
}

func (p *Schema) toType(c schema.ColumnType, co schema.ColumnData) string {
	p.modifyColumnOptionFromType(c, co)

	switch c {
	case schema.ColumnTypeString:
		return "TEXT"
	case schema.ColumnTypeText:
		return "TEXT"
	case schema.ColumnTypeInteger:
		return "INTEGER"
	case schema.ColumnTypeSerial:
		return "INTEGER" // SQLite does not have SERIAL, SERIAL is for PRIMARY KEY with the add of AUTOINCREMENT
	case schema.ColumnTypeBigInt:
		return "INTEGER"
	case schema.ColumnTypeFloat:
		return "REAL"
	case schema.ColumnTypeDecimal, schema.ColumnTypeNumeric:
		return "NUMERIC"
	case schema.ColumnTypeDatetime:
		return "DATETIME"
	case schema.ColumnTypeTime:
		return "TEXT" // SQLite does not have a native TIME type
	case schema.ColumnTypeDate:
		return "TEXT" // SQLite does not have a native DATE type
	case schema.ColumnTypeBinary, schema.ColumnTypeBlob:
		return "BLOB"
	case schema.ColumnTypeBoolean:
		return "INTEGER" // SQLite does not have a native BOOLEAN type, typically use INTEGER
	case schema.ColumnTypeUUID:
		return "TEXT" // UUIDs are typically stored as TEXT in SQLite
	case schema.ColumnTypeJSON:
		return "TEXT" // SQLite does not have a native JSON type, store as TEXT
	case schema.ColumnTypeJSONB:
		return "TEXT" // SQLite does not have a native JSONB type, store as TEXT
	case schema.ColumnTypeHstore:
		return "TEXT" // SQLite does not have a native HSTORE type, store as TEXT

	default:
		return c
	}
}

func (p *Schema) modifyColumnOptionFromType(c schema.ColumnType, co schema.ColumnData) {
	switch c {
	case schema.ColumnTypePrimaryKey:
		co.SetNotNull(true)
		co.SetPrimaryKey(true)
	case schema.ColumnTypeDatetime:
		if co.GetPrecision() == 0 {
			co.SetPrecision(6)
		}
	}
}
