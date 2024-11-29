package amigo

import (
	"context"
	"database/sql"
	"fmt"
)

// GetStatus return the state of the database
func (a Amigo) GetStatus(ctx context.Context, db *sql.DB) ([]string, error) {
	schema, err := a.GetSchema(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("unable to get schema: %w", err)
	}

	return schema.FindAppliedVersions(), nil
}
