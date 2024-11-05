# Running your first migration

Since you have initialized mig, you can now run your first migration.

Before that make sure to import the driver that amigo have added in the imports of the main file in `migrations/db/main.go`.

Here is an example of amigo main: 

```go
package main

import (
	"database/sql"
	migrations "github.com/alexisvisco/gwt/migrations"
	"github.com/alexisvisco/amigo/pkg/entrypoint"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	_ "github.com/jackc/pgx/v5/stdlib" // <- you can switch to any driver that support the database/sql package
	"os"
)

func main() {
	opts, arg := entrypoint.AmigoContextFromFlags()

	db, err := sql.Open("pgx", opts.GetRealDSN()) // <- change this line too, the dsn is what you provided in the parameter or context configuration
	if err != nil {
		logger.Error(events.MessageEvent{Message: err.Error()})
		os.Exit(1)
	}

	entrypoint.Main(db, arg, migrations.Migrations, opts)
}
```

By default for postgres it imports `github.com/jackc/pgx/v5/stdlib` but you can change it and it will works. 

Amigo is driver agnostic and works with the `database/sql` package.
`pgx` provide a support for the `database/sql` package and is a good choice for postgres, but you can use any driver that support the `database/sql` package.


When you have installed the driver on you project, run the migration, execute the following command:

```sh
amigo migrate
```
