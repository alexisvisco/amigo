# Migrating in go

Usually you will need to run migration in your Go application, to do so you can use the `amigo` package.

```go
package main

import (
	"database/sql"
	migrations "package/where/your/migrations/are"
	"github.com/alexisvisco/amigo/pkg/amigo"
	_ "github.com/lib/pq"
)


func main() {
	db, _ := sql.Open("postgres", "postgres://user:password@localhost:5432/dbname?sslmode=disable")
	ok, err := amigo.RunPostgresMigrations(&amigo.RunMigrationOptions{
		Connection: db,
		Migrations: migrations.Migrations,
	})
}
```

You can specify all the options the cli can take in the `RunMigrationOptions` struct (steps, version, dryrun ...)
