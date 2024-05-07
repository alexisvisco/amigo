package main

import (
	migrations "github.com/alexisvisco/mig/migrations"
	"github.com/alexisvisco/mig/pkg/entrypoint"
)

// Main is the entrypoint of the migrations
// It will parse the flags and execute the migrations
// Available flags are:
// - dsn: URL connection to the database
// - version: Migrate or rollback a specific version
// - direction: UP or DOWN
// - json: Print the output in JSON
// - silent: Do not print migrations output
// - timeout: Timeout for the migration is the time for the whole migrations to be applied
// - dry-run: Dry run the migration will not apply the migration to the database
// - continue-on-error: Continue on error will not rollback the migration if an error occurs
// - schema-version-table: Table name for the schema version
// - verbose: Print SQL statements
func main() {
	entrypoint.MainPostgres(migrations.Migrations)
}
