package schema

import (
	"fmt"
	"github.com/gobuffalo/flect"
	"strings"
)

// AddCheckConstraint returns true if the check constraint was added, false if it already exists.
// Adds a new check constraint to the Table. expression is a String representation of verifiable boolean condition.
//
// Example:
//
//	p.AddCheckConstraint("products", "price_check", "price > 0", CheckConstraintOptions{})
//
// Generates:
//
//	ALTER TABLE "products" ADD CONSTRAINT price_check CHECK (price > 0)
func (p *Postgres) AddCheckConstraint(tableName TableName, constraintName string, expression string, options CheckConstraintOptions) {
	options.Table = tableName
	options.Expression = expression
	options.ConstraintName = options.BuildConstraintName(tableName, constraintName)

	if p.Context.migrationType == MigrationTypeDown {
		// todo: implement down migration
		return
	}

	if options.IfNotExists {
		exists := p.ConstraintExist(options.Table, options.ConstraintName)
		if exists {
			return
		}
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD %s", options.Table.String(), p.checkConstraint(options))
	_, err := p.db.ExecContext(p.Context.Context, query)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding check constraint: %w", err))
		return
	}

	p.Context.addCheckConstraintCreated(options)
}

func (p *Postgres) checkConstraint(options CheckConstraintOptions) string {
	query := `CONSTRAINT {constraint_name} CHECK {expression} {validate}`

	replacer := replacer{
		"constraint_name": strfunc(options.ConstraintName),
		"expression":      strfunc(parentheses(options.Expression)),
		"validate": func() string {
			if options.Validate != nil && !*options.Validate {
				return "NOT VALID"
			}
			return ""
		},
	}

	return replacer.replace(query)
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
func (p *Postgres) AddForeignKeyConstraint(fromTable, toTable TableName, options AddForeignKeyConstraintOptions) {
	options.FromTable = fromTable
	options.ToTable = toTable
	options.ForeignKeyName = options.BuildForeignKeyName(options.FromTable, options.ToTable)

	if p.Context.migrationType == MigrationTypeDown {
		// todo: implement down migration
	}

	if options.IfNotExists {
		exists := p.ConstraintExist(options.FromTable, options.ForeignKeyName)
		if exists {
			return
		}
	}

	_, err := p.db.ExecContext(p.Context.Context,
		fmt.Sprintf("ALTER TABLE %s ADD %s", options.FromTable, p.foreignKeyConstraint(options)))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding foreign key: %w", err))
		return
	}

	p.Context.addForeignKeyCreated(options)
}

func (p *Postgres) foreignKeyConstraint(options AddForeignKeyConstraintOptions) string {
	query := `CONSTRAINT {foreign_key_name} FOREIGN KEY ({column}) REFERENCES {to_table} ({primary_key}) {on_delete} {on_update} {deferrable}`

	replacer := replacer{
		"from_table":       strfunc(options.FromTable.String()),
		"to_table":         strfunc(options.ToTable.String()),
		"foreign_key_name": strfunc(options.ForeignKeyName),

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
			// with the option provided by the user
			return options.Deferrable
		},
	}

	return replacer.replace(query)
}

func (p *Postgres) references(tableName TableName, column string) string {
	return fmt.Sprintf("%s (%s)", tableName, column)
}

// AddPrimaryKeyConstraint adds a new primary key to the Table.
//
// Example:
//
//	p.AddPrimaryKeyConstraint("users", []string{"id"}, PrimaryKeyConstraintOptions{})
//
// Generates:
//
//	ALTER TABLE "users" ADD CONSTRAINT pk_users PRIMARY KEY (id)
//
// Adding a composite primary key:
//
//	p.AddPrimaryKeyConstraint("users", []string{"id", "name"}, PrimaryKeyConstraintOptions{})
//
// Generates:
//
//	ALTER TABLE "users" ADD CONSTRAINT pk_users PRIMARY KEY (id, name)
//
// Adding a primary key with a custom name:
//
//	p.AddPrimaryKeyConstraint("users", []string{"id"}, PrimaryKeyConstraintOptions{ConstraintName: "custom_pk_users"})
//
// Generates:
//
//	ALTER TABLE "users" ADD CONSTRAINT custom_pk_users PRIMARY KEY (id)
//
// Adding a primary key if it does not exist:
//
//	p.AddPrimaryKeyConstraint("users", []string{"id"}, PrimaryKeyConstraintOptions{IfNotExists: true})
func (p *Postgres) AddPrimaryKeyConstraint(tableName TableName, columns []string, options PrimaryKeyConstraintOptions) {
	options.Table = tableName
	options.Columns = columns

	if p.Context.migrationType == MigrationTypeDown {
		// todo: implement down migration
	}

	if options.IfNotExists {
		exists := p.PrimaryKeyExists(options.Table)
		if exists {
			return
		}
	}

	sql := fmt.Sprintf("ALTER TABLE %s ADD %s", options.Table, p.primaryKeyConstraint(options))
	_, err := p.db.ExecContext(p.Context.Context, sql)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding primary key: %w", err))
		return
	}

	p.Context.addPrimaryKeyCreated(options)
}

func (p *Postgres) primaryKeyConstraint(options PrimaryKeyConstraintOptions) string {
	sql := `PRIMARY KEY {columns}`

	replacer := replacer{
		"columns": func() string {
			if len(options.Columns) == 0 {
				return ""
			}
			return fmt.Sprintf("(%s)", strings.Join(options.Columns, ", "))
		},
	}

	return replacer.replace(sql)
}

func (p *Postgres) toReferentialAction(action string) string {
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

func (p *Postgres) applyConstraint(opt any) string {
	switch t := opt.(type) {
	case CheckConstraintOptions:
		return p.checkConstraint(t)
	case AddForeignKeyConstraintOptions:
		return p.foreignKeyConstraint(t)
	case PrimaryKeyConstraintOptions:
		return p.primaryKeyConstraint(t)
	default:
		p.Context.RaiseError(fmt.Errorf("unsupported constraint type: %T", t))
	}

	return ""
}
