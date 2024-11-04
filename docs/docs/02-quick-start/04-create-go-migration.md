# Go migration
Now you can create your own migration, to do that you can run the following command:

```sh
amigo create create_users_table
```

This command will create a new migration file `{DATE}_create_users_table.go` in the `migrations` folder.

This migration is in go, you have two kinds of migration files:
- classic: they are up and down migrations as they implement [DetailedMigration](https://github.com/alexisvisco/amigo/blob/main/pkg/schema/migrator.go#L40) interface
- change: they are up migrations as they implement [ChangeMigration](https://github.com/alexisvisco/amigo/blob/main/pkg/schema/migrator.go#L48) interface

(more information on the cli [here](../03-cli/03-create.md))


### Example of a `change` migration file:

```go
package migrations

import (
    "github.com/alexisvisco/amigo/pkg/schema/pg"
)

type Migration20240524090434CreateUserTable struct {}

func (m Migration20240524090434CreateUserTable) Change(s *pg.Schema) {
    s.CreateTable("users", func(def *pg.PostgresTableDef) {
        def.Serial("id")
        def.String("name")
        def.String("email")
        def.Timestamps()
        def.Index([]string{"name"})
    })
}
```

### Example of a `classic` migration file:

```go
package migrations

import (
    "github.com/alexisvisco/amigo/pkg/schema/pg"
)

type Migration20240524090434CreateUserTable struct {}

func (m Migration20240524090434CreateUserTable) Up(s *pg.Schema) {
    s.CreateTable("users", func(def *pg.PostgresTableDef) {
        def.Serial("id")
        def.String("name")
        def.String("email")
        def.Timestamps()
        def.Index([]string{"name"})
    })
}

func (m Migration20240524090434CreateUserTable) Down(s *pg.Schema) {
    s.DropTable("users")
}
```


