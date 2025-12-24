package amigo

import (
	"context"
	"database/sql"
	"time"
)

type Configuration struct {
	// Directory is the location of the migrations files
	Directory string

	// DB is the database connection
	DB *sql.DB

	// Driver is the database driver for amigo, it's used to create the schema_migrations table and track
	// applied migrations
	Driver Driver

	// SQLFileUpAnnotation is the annotation used to indicate the start of the up migration in a SQL file
	SQLFileUpAnnotation string

	// SQLFileDownAnnotation is the annotation used to indicate the start of the down migration in a SQL file
	SQLFileDownAnnotation string

	// DefaultTransactional indicates if new migrations should be run inside a transaction by wrapping them in a Tx helper
	// or putting the tx annotation in SQL files
	DefaultTransactional bool

	// DefaultFileFormat is the default file format for new migrations (sql or go)
	DefaultFileFormat string
}

var DefaultConfiguration = Configuration{
	Directory: "db/migrations",

	SQLFileDownAnnotation: "-- migrate:down",
	SQLFileUpAnnotation:   "-- migrate:up",

	DefaultTransactional: true,
	DefaultFileFormat:    "sql",
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
}

type MigrationStatus struct {
	Migration MigrationRecord
	Applied   bool
}
