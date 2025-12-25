package amigo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type SQLiteDriver struct {
	tableName string
}

func NewSQLiteDriver(tableName string) *SQLiteDriver {
	if tableName == "" {
		tableName = "schema_migrations"
	}
	return &SQLiteDriver{tableName: tableName}
}

func (d *SQLiteDriver) CreateSchemaMigrationsTableIfNotExists(ctx context.Context, db *sql.DB) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			date INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT (DATETIME('now', 'utc'))
		)
	`, d.tableName)

	_, err := db.ExecContext(ctx, query)
	return err
}

func (d *SQLiteDriver) GetAppliedMigrations(ctx context.Context, db *sql.DB) ([]MigrationRecord, error) {
	query := fmt.Sprintf(`SELECT date, name, applied_at FROM %s ORDER BY date ASC`, d.tableName)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []MigrationRecord
	for rows.Next() {
		var m MigrationRecord
		if err := rows.Scan(&m.Date, &m.Name, &m.AppliedAt); err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return migrations, rows.Err()
}

func (d *SQLiteDriver) InsertMigrations(ctx context.Context, db *sql.DB, list []MigrationRecord) error {
	if len(list) == 0 {
		return nil
	}

	placeholders := make([]string, len(list))
	args := make([]any, 0, len(list)*2)
	for i, m := range list {
		placeholders[i] = "(?, ?)"
		args = append(args, m.Date, m.Name)
	}

	query := fmt.Sprintf(`INSERT INTO %s (date, name) VALUES %s`, d.tableName, strings.Join(placeholders, ", "))
	_, err := db.ExecContext(ctx, query, args...)
	return err
}

func (d *SQLiteDriver) DeleteMigrations(ctx context.Context, db *sql.DB, dates []int64) error {
	if len(dates) == 0 {
		return nil
	}

	placeholders := make([]string, len(dates))
	args := make([]any, len(dates))
	for i, date := range dates {
		placeholders[i] = "?"
		args[i] = date
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE date IN (%s)`, d.tableName, strings.Join(placeholders, ", "))
	_, err := db.ExecContext(ctx, query, args...)
	return err
}

func (d *SQLiteDriver) Name() string {
	return "sqlite"
}
