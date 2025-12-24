package main

import (
	"database/sql"
	"example/migrations"
	"log"
	"os"

	"github.com/alexisvisco/amigo"

	_ "modernc.org/sqlite"
)

func main() {
	// Open SQLite database file
	dbPath := "example.db"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create driver
	driver := amigo.NewSQLiteDriver("schema_migrations")

	// Setup configuration
	config := amigo.Configuration{
		Directory:             "migrations",
		DB:                    db,
		Driver:                driver,
		DebugSQL:              false,
		SQLFileUpAnnotation:   "-- migrate:up",
		SQLFileDownAnnotation: "-- migrate:down",
		DefaultTransactional:  true,
		DefaultFileFormat:     "sql",
	}

	// Load migrations from migrations package
	migrationList := migrations.Migrations(config)

	// Create CLI
	cli := amigo.NewCLI(amigo.CLIConfig{
		Config:     config,
		Migrations: migrationList,
	})

	// Run CLI
	os.Exit(cli.Run(os.Args[1:]))
}
