package base

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
)

// Schema is the base schema. It is used to support unknown database types and provide a default implementation.
type Schema struct {
	// TX is the transaction to execute the queries.
	TX schema.DB

	// DB is a database connection but not in a transaction.
	DB schema.DB

	Context *schema.MigratorContext

	// ReversibleMigrationExecutor is a helper to execute reversible migrations in change method.
	*schema.ReversibleMigrationExecutor
}

// NewBase creates a new base schema.
func NewBase(ctx *schema.MigratorContext, tx schema.DB, db schema.DB) *Schema {
	return &Schema{
		TX:                          tx,
		DB:                          db,
		Context:                     ctx,
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
		ReversibleMigrationExecutor: schema.NewReversibleMigrationExecutor(&ctx),
	}
}

// AddVersion adds a new version to the schema_migrations table.
// This function is not reversible.
func (p *Schema) AddVersion(version string) {
	sql := `INSERT INTO {version_table} (id) VALUES ({version})`

	replacer := utils.Replacer{
		"version_table": utils.StrFunc(p.Context.MigratorOptions.SchemaVersionTable.String()),
		"version":       utils.StrFunc(fmt.Sprintf("'%s'", version)),
	}

	_, err := p.TX.ExecContext(p.Context.Context, replacer.Replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding version: %w", err))
		return
	}

	p.Context.AddVersionCreated(version)
}

// RemoveVersion removes a version from the schema_migrations table.
// This function is not reversible.
func (p *Schema) RemoveVersion(version string) {
	sql := `DELETE FROM {version_table} WHERE id = {version}`

	replacer := utils.Replacer{
		"version_table": utils.StrFunc(p.Context.MigratorOptions.SchemaVersionTable.String()),
		"version":       utils.StrFunc(fmt.Sprintf("'%s'", version)),
	}

	_, err := p.TX.ExecContext(p.Context.Context, replacer.Replace(sql))
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

	rows, err := p.TX.QueryContext(p.Context.Context, replacer.Replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while fetching applied versions: %w", err))
		return nil
	}

	defer rows.Close()

	var versions []string

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			p.Context.RaiseError(fmt.Errorf("error while scanning version: %w", err))
			return nil
		}
		versions = append(versions, version)
	}

	if err := rows.Err(); err != nil {
		p.Context.RaiseError(fmt.Errorf("error after iterating rows: %w", err))
		return nil
	}

	return versions
}
