package amigo

import (
	"database/sql"
	"fmt"
)

func (a Amigo) SkipMigrationFile(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO "+a.ctx.SchemaVersionTable+" (id) VALUES ($1)", a.ctx.Create.Version)
	if err != nil {
		return fmt.Errorf("unable to skip migration file: %w", err)
	}

	return nil
}
