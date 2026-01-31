package main

import (
	"database/sql"
	"example-postgres/migrations"
	"log"
	"os"

	"github.com/alexisvisco/amigo"
	"github.com/alexisvisco/amigo/pkg/logsql"

	_ "github.com/lib/pq"
)

func main() {
	// Wrap driver with debug logging
	driverName := logsql.WrapDriver("postgres", logsql.WrapOptionOutput(os.Stdout))

	// Connection string for docker postgres
	dsn := "postgres://postgres:postgres@localhost:5432/amigo_test?sslmode=disable"
	if envDSN := os.Getenv("DATABASE_URL"); envDSN != "" {
		dsn = envDSN
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create driver
	driver := amigo.NewPostgresDriver("schema_migrations")

	// Setup configuration
	config := amigo.DefaultConfiguration
	config.DB = db
	config.Driver = driver
	config.SplitStatements = true

	// Load migrations from migrations package
	migrationList := migrations.Migrations(config)

	// Create CLI
	cli := amigo.NewCLI(amigo.CLIConfig{
		Config:               config,
		Migrations:           migrationList,
		Directory:            "migrations",
		DefaultTransactional: true,
		DefaultFileFormat:    "sql",
	})

	// Run CLI
	os.Exit(cli.Run(os.Args[1:]))
}
