package pg

import (
	"fmt"

	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/base"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
)

type Schema struct {
	// TX is the transaction to execute the queries.
	TX schema.DB

	// DB is a database connection but not in a transaction.
	DB schema.DB

	Context *schema.MigratorContext

	*base.Schema

	// ReversibleMigrationExecutor is a helper to execute reversible migrations in change method.
	*schema.ReversibleMigrationExecutor
}

func NewPostgres(ctx *schema.MigratorContext, tx schema.DB, db schema.DB) *Schema {
	return &Schema{
		TX:                          tx,
		DB:                          db,
		Context:                     ctx,
		Schema:                      base.NewBase(ctx, tx, db),
		ReversibleMigrationExecutor: schema.NewReversibleMigrationExecutor(ctx),
	}
}

// rollbackMode will allow to execute migration without getting a infinite loop by checking the migration direction.
func (p *Schema) rollbackMode() *Schema {
	ctx := *p.Context
	ctx.MigrationDirection = types.MigrationDirectionNotReversible
	return &Schema{
		TX:                          p.TX,
		DB:                          p.DB,
		Context:                     &ctx,
		Schema:                      base.NewBase(&ctx, p.TX, p.DB),
		ReversibleMigrationExecutor: schema.NewReversibleMigrationExecutor(&ctx),
	}
}

func (p *Schema) Exec(query string, args ...interface{}) {
	_, err := p.TX.ExecContext(p.Context.Context, query, args...)
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
		extensionOptions := schema.DropExtensionOptions{IfExists: true}
		if options.Reversible != nil {
			extensionOptions.Cascade = options.Reversible.Cascade
		}
		p.rollbackMode().DropExtension(options.ExtensionName, extensionOptions)
		return
	}

	sql := `CREATE EXTENSION {if_not_exists} "{name}" {schema}`

	replacer := utils.Replacer{
		"if_not_exists": utils.StrFuncPredicate(options.IfNotExists, "IF NOT EXISTS"),
		"name":          utils.StrFunc(options.ExtensionName),
		"schema":        utils.StrFuncPredicate(options.Schema != "", fmt.Sprintf("SCHEMA %s", options.Schema)),
	}

	_, err := p.TX.ExecContext(p.Context.Context, replacer.Replace(sql))
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

	sql := `DROP EXTENSION {if_exists} "{name}" {cascade}`

	replacer := utils.Replacer{
		"if_exists": utils.StrFuncPredicate(options.IfExists, "IF EXISTS"),
		"name":      utils.StrFunc(p.toExtension(options.ExtensionName)),
		"cascade":   utils.StrFuncPredicate(options.Cascade, "CASCADE"),
	}

	_, err := p.TX.ExecContext(p.Context.Context, replacer.Replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while dropping extension: %w", err))
		return
	}

	p.Context.AddExtensionDropped(options)
}

func (p *Schema) toExtension(extension string) string {
	switch extension {
	case "uuid":
		return "uuid-ossp"
	default:
		return extension
	}
}
