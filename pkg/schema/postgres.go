package schema

import (
	"fmt"
)

// https://api.rubyonrails.org/classes/ActiveRecord/ConnectionAdapters/Table.html#method-i-change

/*
A
add_column, add_foreign_key, add_index, add_reference, add_timestamps, assume_migrated_upto_version
B
build_create_table_definition
C
change_column, change_column_comment, change_column_default, change_column_null, change_table, change_table_comment, check_constraint_exists?, check_constraints, column_exists?, Columns, create_join_table, create_table
D
data_source_exists?, data_sources, drop_join_table, drop_table
F
foreign_key_exists?, foreign_keys
I
index_exists?, index_name_exists?, indexes
M
max_index_name_size
N
native_database_types
O
options_include_default?
P
primary_key
R
remove_belongs_to, remove_check_constraint, remove_column, remove_columns, remove_foreign_key, remove_index, remove_reference, remove_timestamps, rename_column, rename_index, rename_table
T
table_alias_for, table_comment, table_exists?, table_options, tables
U
use_foreign_keys?
V
view_exists?, views
*/

type Postgres struct {
	db      DB
	Context MigratorContext
	ReversibleMigrationExec
}

func NewPostgres(ctx MigratorContext, db DB) *Postgres {
	return &Postgres{db: db, Context: ctx, ReversibleMigrationExec: ReversibleMigrationExec{ctx}}
}

// AddExtension adds a new extension to the database.
//
// Example:
//
//	p.AddExtension("uuid", ExtensionOptions{})
//
// Generates:
//
//	CREATE EXTENSION IF NOT EXISTS "uuid-ossp"
func (p *Postgres) AddExtension(name string, options ExtensionOptions) {
	options.ExtensionName = p.toExtension(name)

	if p.Context.migrationType == MigrationTypeDown {
		p.DropExtension(options.ExtensionName, DropExtensionOptions{IfExists: true})
		return
	}

	p.executeAddExtension(options)
}

func (p *Postgres) executeAddExtension(options ExtensionOptions) {
	sql := `CREATE EXTENSION {if_not_exists} "{name}" {schema}`

	replacer := replacer{
		"if_not_exists": func() string {
			if options.IfNotExists {
				return "IF NOT EXISTS"
			}
			return ""
		},

		"name": strfunc(options.ExtensionName),
		"schema": func() string {
			if options.Schema != "" {
				return fmt.Sprintf("SCHEMA %s", options.Schema)
			}
			return ""
		},
	}

	_, err := p.db.ExecContext(p.Context.Context, replacer.replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding extension: %w", err))
		return
	}

	p.Context.addExtensionCreated(options)
}

// DropExtension drops an extension from the database.
//
// Example:
//
//	p.DropExtension("uuid", DropExtensionOptions{})
//
// Generates:
//
//	DROP EXTENSION IF EXISTS "uuid-ossp"
func (p *Postgres) DropExtension(name string, options DropExtensionOptions) {
	options.ExtensionName = name

	if p.Context.migrationType == MigrationTypeDown && options.Reversible != nil {
		p.executeAddExtension(*options.Reversible)
		return
	}

	p.executeDropExtension(options)
}

func (p *Postgres) executeDropExtension(options DropExtensionOptions) {
	sql := `DROP EXTENSION {if_exists} "{name}"`

	replacer := replacer{
		"if_exists": func() string {
			if options.IfExists {
				return "IF EXISTS"
			}
			return ""
		},

		"name": strfunc(p.toExtension(options.ExtensionName)),
	}

	_, err := p.db.ExecContext(p.Context.Context, replacer.replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while dropping extension: %w", err))
		return
	}

	p.Context.addExtensionDropped(options)
}

func (p *Postgres) toExtension(extension string) string {
	switch extension {
	case "uuid":
		return "uuid-ossp"
	default:
		return extension
	}
}
