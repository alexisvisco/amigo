package amigo

import (
	"context"
	"database/sql"

	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/base"
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"github.com/alexisvisco/amigo/pkg/schema/sqlite"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils/dblog"
	sqldblogger "github.com/simukti/sqldb-logger"
)

type Amigo struct {
	Config              *amigoconfig.Config
	Driver              types.Driver
	CustomSchemaFactory schema.Factory[schema.Schema]
}

type OptionFn func(a *Amigo)

func WithCustomSchemaFactory(factory schema.Factory[schema.Schema]) OptionFn {
	return func(a *Amigo) {
		a.CustomSchemaFactory = factory
	}
}

// NewAmigo create a new amigo instance
func NewAmigo(ctx *amigoconfig.Config, opts ...OptionFn) Amigo {
	a := Amigo{
		Config: ctx,
		Driver: types.GetDriver(ctx.DSN),
	}

	for _, opt := range opts {
		opt(&a)
	}

	return a
}

type MigrationApplier interface {
	Apply(direction types.MigrationDirection, version *string, steps *int, migrations []schema.Migration) bool
	GetSchema() schema.Schema
}

func (a Amigo) GetMigrationApplier(
	ctx context.Context,
	conn *sql.DB,
) (MigrationApplier, error) {
	recorder := dblog.NewHandler(a.Config.ShowSQLSyntaxHighlighting)
	recorder.ToggleLogger(true)

	if a.Config.ValidateDSN() == nil {
		conn = sqldblogger.OpenDriver(a.Config.GetRealDSN(), conn.Driver(), recorder)
	}

	if a.CustomSchemaFactory != nil {
		return schema.NewMigrator(ctx, conn, a.CustomSchemaFactory, a.Config), nil
	}

	switch a.Driver {
	case types.DriverPostgres:
		return schema.NewMigrator(ctx, conn, pg.NewPostgres, a.Config), nil
	case types.DriverSQLite:
		return schema.NewMigrator(ctx, conn, sqlite.NewSQLite, a.Config), nil
	}

	return schema.NewMigrator(ctx, conn, base.NewBase, a.Config), nil
}

func (a Amigo) GetSchema(ctx context.Context, conn *sql.DB) (schema.Schema, error) {
	applier, err := a.GetMigrationApplier(ctx, conn)
	if err != nil {
		return nil, err
	}

	return applier.GetSchema(), nil
}
