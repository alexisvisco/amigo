# Migrating in go

Usually you will need to run migration in your Go application, to do so you can use the `amigo` package.

```go
package main

import (
	"database/sql"
	"example/pg/migrations"
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoctx"
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

	err = amigo.NewAmigo(amigoctx.NewContext().WithDSN(dsn)).RunMigrations(amigo.RunMigrationParams{
		DB:         db,
		Direction:  types.MigrationDirectionDown,
		Migrations: migrations.Migrations,
		LogOutput:  os.Stdout,
	})
	if err != nil {
		panic(err)
	}
}
```

You can specify all the options the cli can take in the `RunMigrationOptions` struct (steps, version, dryrun ...)
