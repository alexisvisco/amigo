# Migrating in go

Usually you will need to run migration in your Go application, to do so you can use the `amigo` package.

```go
package main

import (
	"database/sql"
	"example/pg/db/migrations"
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/types"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
)

// this is an example to run migration in a codebase
func main() {
	dsn := "postgres://postgres:postgres@localhost:6666/postgres"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}

	err = migrateDatabase(dsn, db) 
	if err != nil {
		panic(err)
    }
}

func migrateDatabase(databaseURL string, rawDB *sql.DB) error {
	dumpSchemaAfterMigrating := os.Getenv("DUMP_SCHEMA_AFTER_MIGRATING") == "true"

	a := amigoctx.NewContext().
		WithDSN(databaseURL). // you need to provide the dsn too, in order for amigo to detect the driver
		WithShowSQL(true). // will show you sql queries
		WithDumpSchemaAfterMigrating(dumpSchemaAfterMigrating) // will create/modify the schema.sql in db folder (only in local is suitable)

	err := amigo.NewAmigo(a).RunMigrations(amigo.RunMigrationParams{
		DB:         rawDB,
		Direction:  amigotypes.MigrationDirectionUp, // will migrate the database up
		Migrations: migrations.Migrations, // where migrations are located (db/migrations)
		Logger:     slog.Default(), // you can also only specify the LogOutput and it will use the default amigo logger at your desired output (io.Writer)
	})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}
```

You can specify all the options the cli can take in the `RunMigrationOptions` struct (steps, version, dryrun ...)
