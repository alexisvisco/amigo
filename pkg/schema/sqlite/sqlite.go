package sqlite

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/base"
	"github.com/alexisvisco/amigo/pkg/types"
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

func NewSQLite(ctx *schema.MigratorContext, tx schema.DB, db schema.DB) *Schema {
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
