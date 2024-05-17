package pg

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/gobuffalo/flect"
	"strings"
)

// AddCheckConstraint Adds a new check constraint to the Table.
// expression parameter is a FormatRecords representation of verifiable boolean condition.
//
// Example:
//
//	p.AddCheckConstraint("products", "price_check", "price > 0")
//
// Generates:
//
//	ALTER TABLE "products" ADD CONSTRAINT price_check CHECK (price > 0)
func (p *Schema) AddCheckConstraint(tableName schema.TableName, constraintName string, expression string, opts ...schema.CheckConstraintOptions) {
	options := schema.CheckConstraintOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Table = tableName
	options.Expression = expression
	options.ConstraintName = options.BuildConstraintName(tableName, constraintName)

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		p.rollbackMode().DropCheckConstraint(tableName, constraintName,
			schema.DropCheckConstraintOptions{IfExists: true})
		return
	}

	if options.IfNotExists {
		exists := p.ConstraintExist(options.Table, options.ConstraintName)
		if exists {
			return
		}
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD %s", options.Table.String(), p.checkConstraint(options))
	_, err := p.DB.ExecContext(p.Context.Context, query)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding check constraint: %w", err))
		return
	}

	p.Context.AddCheckConstraintCreated(options)
}

func (p *Schema) checkConstraint(options schema.CheckConstraintOptions) string {
	query := `CONSTRAINT {constraint_name} CHECK {expression} {validate}`

	replacer := utils.Replacer{
		"constraint_name": utils.StrFunc(options.ConstraintName),
		"expression":      utils.StrFunc(utils.Parentheses(options.Expression)),
		"validate": func() string {
			if options.Validate != nil && !*options.Validate {
				return "NOT VALID"
			}
			return ""
		},
	}

	return replacer.Replace(query)
}

// DropCheckConstraint drops a check constraint from the Table.
//
// Example:
//
//	p.DropCheckConstraint("products", "price_check")
//
// Generates:
//
//	ALTER TABLE "products" DROP CONSTRAINT price_check
//
// Dropping a check constraint if it exists:
//
//	p.DropCheckConstraint("products", "price_check", schema.DropCheckConstraintOptions{IfExists: true})
//
// Generates:
//
//	ALTER TABLE "products" DROP CONSTRAINT IF EXISTS price_check
//
// Dropping a check constraint with a reversible expression:
//
//	p.DropCheckConstraint("products", "price_check", schema.DropCheckConstraintOptions{
//		Reversible: schema.CheckConstraintOptions{Expression: "price > 0"},
//	})
//
// Generates:
//
//	ALTER TABLE "products" ADD CONSTRAINT price_check CHECK (price > 0)
func (p *Schema) DropCheckConstraint(tableName schema.TableName, constraintName string, opts ...schema.DropCheckConstraintOptions) {
	options := schema.DropCheckConstraintOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Table = tableName
	options.ConstraintName = options.BuildConstraintName(tableName, constraintName)

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		if options.Reversible != nil {
			p.rollbackMode().AddCheckConstraint(tableName, constraintName, options.Reversible.Expression,
				*options.Reversible)
		} else {
			logger.Warn(events.MessageEvent{
				Message: fmt.Sprintf("unable re-creating  check constraint %s", options.ConstraintName),
			})
		}
		return
	}

	err := p.dropConstraint(tableName, options.ConstraintName, options.IfExists)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while dropping check constraint: %w", err))
		return
	}

	p.Context.AddCheckConstraintDropped(options)
}

// AddForeignKeyConstraint
// Adds a new foreign key. FromTable is the Table with the key column, ToTable contains the referenced primary key.
//
// The foreign key will be named after the following pattern: fk_[from_table]_[to_table].
//
// Creating a simple foreign key:
//
//	p.AddForeignKeyConstraint("articles", "authors", AddForeignKeyConstraintOptions{})
//
// Generates:
//
//	ALTER TABLE "articles" ADD CONSTRAINT fk_articles_authors FOREIGN KEY (author_id) REFERENCES "authors" (id)
//
// Creating a foreign key, ignoring method call if the foreign key exists:
//
//	p.AddForeignKeyConstraint("articles", "authors", AddForeignKeyConstraintOptions{IfNotExists: true})
//
// Creating a foreign key on a specific column:
//
//	p.AddForeignKeyConstraint("articles", "users", AddForeignKeyConstraintOptions{Column: "author_id", PrimaryKey: "lng_id"})
//
// Generates:
//
//	ALTER TABLE "articles" ADD CONSTRAINT fk_articles_users FOREIGN KEY (author_id) REFERENCES "users" (lng_id)
//
// Creating a composite foreign key:
//
//	Assuming "carts" Table has "(shop_id, user_id)" as a primary key.
//	p.AddForeignKeyConstraint("orders", "carts", AddForeignKeyConstraintOptions{PrimaryKey: []string{"shop_id", "user_id"}})
//
// Generates:
//
//	ALTER TABLE "orders" ADD CONSTRAINT fk_orders_carts FOREIGN KEY (cart_shop_id, cart_user_id) REFERENCES "carts" (shop_id, user_id)
//
// Creating a cascading foreign key:
//
//	p.AddForeignKeyConstraint("articles", "authors", AddForeignKeyConstraintOptions{OnDelete: "cascade"})
//
// Generates:
//
//	ALTER TABLE "articles" ADD CONSTRAINT fk_articles_authors FOREIGN KEY (author_id) REFERENCES "authors" (id) ON DELETE CASCADE
func (p *Schema) AddForeignKeyConstraint(fromTable, toTable schema.TableName, opts ...schema.AddForeignKeyConstraintOptions) {
	options := schema.AddForeignKeyConstraintOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.FromTable = fromTable
	options.ToTable = toTable
	options.ForeignKeyName = options.BuildForeignKeyName(options.FromTable, options.ToTable)

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		// todo: implement down migration
	}

	if options.IfNotExists {
		exists := p.ConstraintExist(options.FromTable, options.ForeignKeyName)
		if exists {
			return
		}
	}

	_, err := p.DB.ExecContext(p.Context.Context,
		fmt.Sprintf("ALTER TABLE %s ADD %s", options.FromTable, p.foreignKeyConstraint(options)))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding foreign key: %w", err))
		return
	}

	p.Context.AddForeignKeyCreated(options)
}

func (p *Schema) foreignKeyConstraint(options schema.AddForeignKeyConstraintOptions) string {
	query := `CONSTRAINT {foreign_key_name} FOREIGN KEY ({column}) REFERENCES {to_table} ({primary_key}) {on_delete} {on_update} {deferrable}`

	replacer := utils.Replacer{
		"from_table":       utils.StrFunc(options.FromTable.String()),
		"to_table":         utils.StrFunc(options.ToTable.String()),
		"foreign_key_name": utils.StrFunc(options.ForeignKeyName),

		"primary_key": func() string {
			if options.CompositePrimaryKey != nil {
				return fmt.Sprintf("%s", strings.Join(options.CompositePrimaryKey, ", "))
			}

			if options.PrimaryKey != "" {
				return fmt.Sprintf("%s", options.PrimaryKey)
			}

			return "id"
		},

		"column": func() string {
			if options.CompositePrimaryKey != nil {
				prefixed := make([]string, len(options.CompositePrimaryKey))
				for i, v := range options.CompositePrimaryKey {
					prefixed[i] = flect.Singularize(options.ToTable.Name()) + "_" + v
				}
				return fmt.Sprintf("%s", strings.Join(prefixed, ", "))
			}

			if options.Column == "" {
				return fmt.Sprintf("%s_id", flect.Singularize(options.ToTable.Name()))
			}

			return fmt.Sprintf("%s", options.Column)
		},

		"on_delete": func() string {
			if options.OnDelete != "" {
				return fmt.Sprintf("ON DELETE %s", p.toReferentialAction(options.OnDelete))
			}
			return ""
		},

		"on_update": func() string {
			if options.OnUpdate != "" {
				return fmt.Sprintf("ON UPDATE %s", p.toReferentialAction(options.OnUpdate))
			}
			return ""
		},

		"deferrable": func() string {
			// since deferrable options are at the end of the query, we can safely replace the variable
			// with the option provided by the User
			return options.Deferrable
		},
	}

	return replacer.Replace(query)
}

func (p *Schema) references(tableName schema.TableName, column string) string {
	return fmt.Sprintf("%s (%s)", tableName, column)
}

// DropForeignKeyConstraint drops a foreign key from the Table.
//
// Example:
//
//	p.DropForeignKeyConstraint("articles", "authors")
//
// Generates:
//
//	ALTER TABLE "articles" DROP CONSTRAINT fk_articles_authors
//
// Dropping a foreign key if it exists:
//
//	p.DropForeignKeyConstraint("articles", "authors", schema.DropForeignKeyConstraintOptions{IfExists: true})
//
// Generates:
//
//	ALTER TABLE "articles" DROP CONSTRAINT IF EXISTS fk_articles_authors
//
// Dropping a foreign key with a reversible expression:
//
//	p.DropForeignKeyConstraint("articles", "authors", schema.DropForeignKeyConstraintOptions{
//		Reversible: schema.AddForeignKeyConstraintOptions{Column: "author_id", PrimaryKey: "lng_id"},
//	})
//
// Generates:
//
//	ALTER TABLE "articles" ADD CONSTRAINT fk_articles_authors FOREIGN KEY (author_id) REFERENCES "authors" (lng_id)
func (p *Schema) DropForeignKeyConstraint(from, to schema.TableName, opt ...schema.DropForeignKeyConstraintOptions) {
	options := schema.DropForeignKeyConstraintOptions{}
	if len(opt) > 0 {
		options = opt[0]
	}

	options.FromTable = from
	options.ToTable = to
	options.ForeignKeyName = options.BuildForeignKeyName(options.FromTable, options.ToTable)

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		if options.Reversible != nil {
			p.rollbackMode().AddForeignKeyConstraint(from, to, *options.Reversible)
		} else {
			logger.Error(events.MessageEvent{
				Message: fmt.Sprintf("unable re-creating foreign key %s", options.ForeignKeyName),
			})
		}
		return
	}

	err := p.dropConstraint(from, options.ForeignKeyName, options.IfExists)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while dropping foreign key: %w", err))
		return
	}

	p.Context.AddForeignKeyConstraintDropped(options)
}

func (p *Schema) dropConstraint(table schema.TableName, constraintName string, ifExists bool) error {
	query := `ALTER TABLE {table_name} DROP CONSTRAINT {if_exists} {constraint_name}`

	replacer := utils.Replacer{
		"table_name":      utils.StrFunc(table.String()),
		"if_exists":       utils.StrFuncPredicate(ifExists, "IF EXISTS"),
		"constraint_name": utils.StrFunc(constraintName),
	}

	_, err := p.DB.ExecContext(p.Context.Context, replacer.Replace(query))
	if err != nil {
		return fmt.Errorf("error while dropping foreign key: %w", err)
	}
	return nil
}

// AddPrimaryKeyConstraint adds a new primary key to the Table.
//
// Example:
//
//	p.AddPrimaryKeyConstraint("users", []string{"id"}, PrimaryKeyConstraintOptions{})
//
// Generates:
//
//	ALTER TABLE "users" ADD CONSTRAINT PRIMARY KEY (id)
//
// Adding a composite primary key:
//
//	p.AddPrimaryKeyConstraint("users", []string{"id", "name"}, PrimaryKeyConstraintOptions{})
//
// Generates:
//
//	ALTER TABLE "users" ADD CONSTRAINT PRIMARY KEY (id, name)
//
// Adding a primary key if it does not exist:
//
//	p.AddPrimaryKeyConstraint("users", []string{"id"}, PrimaryKeyConstraintOptions{IfNotExists: true})
//
// Generates:
//
//	ALTER TABLE "users" ADD CONSTRAINT IF NOT EXISTS PRIMARY KEY (id)
func (p *Schema) AddPrimaryKeyConstraint(tableName schema.TableName, columns []string, opts ...schema.PrimaryKeyConstraintOptions) {
	options := schema.PrimaryKeyConstraintOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Table = tableName
	options.Columns = columns

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		p.rollbackMode().DropPrimaryKeyConstraint(tableName, schema.DropPrimaryKeyConstraintOptions{IfExists: true})
		return
	}

	if options.IfNotExists {
		exists := p.PrimaryKeyExist(options.Table)
		if exists {
			return
		}
	}

	sql := fmt.Sprintf("ALTER TABLE %s ADD %s", options.Table, p.primaryKeyConstraint(options))
	_, err := p.DB.ExecContext(p.Context.Context, sql)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding primary key: %w", err))
		return
	}

	p.Context.AddPrimaryKeyCreated(options)
}

func (p *Schema) primaryKeyConstraint(options schema.PrimaryKeyConstraintOptions) string {
	sql := `PRIMARY KEY {Columns}`

	replacer := utils.Replacer{
		"Columns": func() string {
			if len(options.Columns) == 0 {
				return ""
			}
			return fmt.Sprintf("(%s)", strings.Join(options.Columns, ", "))
		},
	}

	return replacer.Replace(sql)
}

// DropPrimaryKeyConstraint drops a primary key from the Table.
//
// Example:
//
//	p.DropPrimaryKeyConstraint("users")
//
// Generates:
//
//	ALTER TABLE "users" DROP CONSTRAINT pk_users
//
// Dropping a primary key if it exists:
//
//	p.DropPrimaryKeyConstraint("users", schema.DropPrimaryKeyConstraintOptions{IfExists: true})
//
// Generates:
//
//	ALTER TABLE "users" DROP CONSTRAINT IF EXISTS pk_users
//
// Dropping a primary key with a reversible expression:
//
//	p.DropPrimaryKeyConstraint("users", schema.DropPrimaryKeyConstraintOptions{
//		Reversible: schema.PrimaryKeyConstraintOptions{Columns: []string{"id"}},
//	})
//
// Generates:
//
//	ALTER TABLE "users" ADD CONSTRAINT pk_users PRIMARY KEY (id)
func (p *Schema) DropPrimaryKeyConstraint(tableName schema.TableName, opts ...schema.DropPrimaryKeyConstraintOptions) {
	options := schema.DropPrimaryKeyConstraintOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Table = tableName
	options.PrimaryKeyName = options.BuildPrimaryKeyName(tableName)

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		if options.Reversible != nil {
			p.rollbackMode().AddPrimaryKeyConstraint(tableName, options.Reversible.Columns, *options.Reversible)
		} else {
			logger.Error(events.MessageEvent{
				Message: fmt.Sprintf("unable re-creating primary key %s", options.PrimaryKeyName),
			})
		}
		return
	}

	err := p.dropConstraint(tableName, options.PrimaryKeyName, options.IfExists)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while dropping primary key: %w", err))
		return
	}

	p.Context.AddPrimaryKeyConstraintDropped(options)
}

func (p *Schema) toReferentialAction(action string) string {
	switch strings.ToLower(action) {
	case "cascade":
		action = "CASCADE"
	case "restrict":
		action = "RESTRICT"
	case "nullify":
		action = "SET NULL"
	}

	return action
}

func (p *Schema) applyConstraint(opt any) string {
	switch t := opt.(type) {
	case schema.CheckConstraintOptions:
		return p.checkConstraint(t)
	case schema.AddForeignKeyConstraintOptions:
		return p.foreignKeyConstraint(t)
	case schema.PrimaryKeyConstraintOptions:
		return p.primaryKeyConstraint(t)
	default:
		p.Context.RaiseError(fmt.Errorf("unsupported constraint type: %T", t))
	}

	return ""
}
