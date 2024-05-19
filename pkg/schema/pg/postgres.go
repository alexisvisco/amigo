package pg

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/georgysavva/scany/v2/dbscan"
)

type Schema struct {
	DB      schema.DB
	Context *schema.MigratorContext
	*schema.ReversibleMigrationExecutor
}

func NewPostgres(ctx *schema.MigratorContext, db schema.DB) *Schema {
	return &Schema{DB: db, Context: ctx, ReversibleMigrationExecutor: schema.NewReversibleMigrationExecutor(ctx)}
}

// rollbackMode will allow to execute migration without getting a infinite loop by checking the migration direction.
func (p *Schema) rollbackMode() *Schema {
	ctx := *p.Context
	ctx.MigrationDirection = types.MigrationDirectionNotReversible
	return &Schema{
		DB:                          p.DB,
		Context:                     &ctx,
		ReversibleMigrationExecutor: schema.NewReversibleMigrationExecutor(&ctx),
	}
}

func (p *Schema) Exec(query string, args ...interface{}) {
	_, err := p.DB.ExecContext(p.Context.Context, query, args...)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while executing query: %w", err))
		return
	}
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
func (p *Schema) AddExtension(name string, option ...schema.ExtensionOptions) {
	options := schema.ExtensionOptions{}
	if len(option) > 0 {
		options = option[0]
	}
	options.ExtensionName = p.toExtension(name)

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		p.rollbackMode().DropExtension(options.ExtensionName, schema.DropExtensionOptions{IfExists: true})
		return
	}

	sql := `CREATE EXTENSION {if_not_exists} "{name}" {schema}`

	replacer := utils.Replacer{
		"if_not_exists": utils.StrFuncPredicate(options.IfNotExists, "IF NOT EXISTS"),
		"name":          utils.StrFunc(options.ExtensionName),
		"schema":        utils.StrFuncPredicate(options.Schema != "", fmt.Sprintf("SCHEMA %s", options.Schema)),
	}

	_, err := p.DB.ExecContext(p.Context.Context, replacer.Replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding extension: %w", err))
		return
	}

	p.Context.AddExtensionCreated(options)
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
//
// Dropping an extension if it exists:
//
//	p.DropExtension("uuid", DropExtensionOptions{IfExists: true})
//
// Generates:
//
//	DROP EXTENSION IF EXISTS "uuid-ossp"
//
// To reverse the operation, you can use the reversible option:
//
//	p.DropExtension("uuid", DropExtensionOptions{
//		Reversible: &schema.ExtensionOptions{}
//	})
//
// Generates:
//
//	CREATE EXTENSION "uuid-ossp"
func (p *Schema) DropExtension(name string, opt ...schema.DropExtensionOptions) {
	options := schema.DropExtensionOptions{}
	if len(opt) > 0 {
		options = opt[0]
	}
	options.ExtensionName = name

	if p.Context.MigrationDirection == types.MigrationDirectionDown && options.Reversible != nil {
		p.rollbackMode().AddExtension(name, schema.ExtensionOptions{IfNotExists: true})
		return
	}

	sql := `DROP EXTENSION {if_exists} "{name}"`

	replacer := utils.Replacer{
		"if_exists": utils.StrFuncPredicate(options.IfExists, "IF EXISTS"),
		"name":      utils.StrFunc(p.toExtension(options.ExtensionName)),
	}

	_, err := p.DB.ExecContext(p.Context.Context, replacer.Replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while dropping extension: %w", err))
		return
	}

	p.Context.AddExtensionDropped(options)
}

// AddVersion adds a new version to the schema_migrations table.
// This function is not reversible.
func (p *Schema) AddVersion(version string) {
	sql := `INSERT INTO {version_table} (id) VALUES ($1)`

	replacer := utils.Replacer{
		"version_table": utils.StrFunc(p.Context.MigratorOptions.SchemaVersionTable.String()),
	}

	_, err := p.DB.ExecContext(p.Context.Context, replacer.Replace(sql), version)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding version: %w", err))
		return
	}

	p.Context.AddVersionCreated(version)
}

// RemoveVersion removes a version from the schema_migrations table.
// This function is not reversible.
func (p *Schema) RemoveVersion(version string) {
	sql := `DELETE FROM {version_table} WHERE id = $1`

	replacer := utils.Replacer{
		"version_table": utils.StrFunc(p.Context.MigratorOptions.SchemaVersionTable.String()),
	}

	_, err := p.DB.ExecContext(p.Context.Context, replacer.Replace(sql), version)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while removing version: %w", err))
		return
	}

	p.Context.AddVersionDeleted(version)
}

// FindAppliedVersions returns all the applied versions in the schema_migrations table.
func (p *Schema) FindAppliedVersions() []string {
	sql := `SELECT id FROM {version_table} ORDER BY id ASC`

	replacer := utils.Replacer{
		"version_table": utils.StrFunc(p.Context.MigratorOptions.SchemaVersionTable.String()),
	}

	rows, err := p.DB.QueryContext(p.Context.Context, replacer.Replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while fetching applied versions: %w", err))
		return nil
	}

	var versions []string
	err = dbscan.ScanAll(&versions, rows)
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while scanning applied versions: %w", err))
		return nil
	}

	return versions
}

func (p *Schema) toExtension(extension string) string {
	switch extension {
	case "uuid":
		return "uuid-ossp"
	default:
		return extension
	}
}
