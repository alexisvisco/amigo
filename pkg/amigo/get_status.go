package amigo

import (
	"database/sql"
	"fmt"
)

// GetStatus return the state of the database
func (a Amigo) GetStatus(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT version FROM " + a.ctx.SchemaVersionTable + " ORDER BY version desc")
	if err != nil {
		return nil, fmt.Errorf("unable to get state: %w", err)
	}

	var state []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("unable to scan state: %w", err)
		}
		state = append(state, id)
	}

	return state, nil
}
