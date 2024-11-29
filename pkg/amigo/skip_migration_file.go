package amigo

import (
	"context"
	"database/sql"
	"fmt"
)

func (a Amigo) SkipMigrationFile(ctx context.Context, db *sql.DB) error {
	schema, err := a.GetSchema(ctx, db)
	if err != nil {
		return fmt.Errorf("unable to get schema: %w", err)
	}

	schema.AddVersion(a.Config.Create.Version)

	return nil
}
