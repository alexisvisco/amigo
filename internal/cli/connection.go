package cli

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func GetConnection(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("dsn is required, example: postgres://user:password@localhost:5432/dbname?sslmode=disable")
	}

	if strings.Contains(dsn, "postgres") {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}

		return db, nil
	}

	return nil, errors.New("unsupported database")
}
