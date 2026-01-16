package amigo

import (
	"context"
	"database/sql"
	"time"
)

type Configuration struct {
	// DB is the database connection
	DB *sql.DB

	// Driver is the database driver for amigo, it's used to create the schema_migrations table and track
	// applied migrations
	Driver Driver

	// SQLFileUpAnnotation is the annotation used to indicate the start of the up migration in a SQL file
	SQLFileUpAnnotation string

	// SQLFileDownAnnotation is the annotation used to indicate the start of the down migration in a SQL file
	SQLFileDownAnnotation string

	// SplitStatements controls whether SQL migrations are split by semicolons.
	// Default false: entire migration sent as single exec (PostgreSQL, SQLite)
	// When true: split by semicolons, respecting -- amigo:statement:begin/end annotations (ClickHouse)
	SplitStatements bool
}

var DefaultConfiguration = Configuration{
	SQLFileUpAnnotation:   "-- migrate:up",
	SQLFileDownAnnotation: "-- migrate:down",
}

type Migration interface {
	Up(ctx context.Context, db *sql.DB) error
	Down(ctx context.Context, db *sql.DB) error
	Name() string
	Date() int64
}

type MigrationRecord struct {
	Date      int64
	Name      string
	AppliedAt time.Time
}

type Driver interface {
	CreateSchemaMigrationsTableIfNotExists(ctx context.Context, db *sql.DB) error
	GetAppliedMigrations(ctx context.Context, db *sql.DB) ([]MigrationRecord, error)
	InsertMigrations(ctx context.Context, db *sql.DB, list []MigrationRecord) error
	DeleteMigrations(ctx context.Context, db *sql.DB, dates []int64) error
	Name() string
}

type MigrationStatus struct {
	Migration MigrationRecord
	Applied   bool
}
