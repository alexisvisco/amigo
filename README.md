# Amigo - Go SQL Migration Tool

A simple, powerful SQL migration tool for Go with support for both SQL and Go migrations.

Amigo provides a clean API for managing database migrations with built-in CLI support and programmatic access. Write migrations in SQL or Go, control transactions, and get real-time feedback during execution.

## Features

- **SQL and Go migrations** - Write migrations in SQL files or Go code
- **Embedded migrations** - SQL files are embedded in binary via `embed.FS` for portability
- **Transaction control** - Fine-grained control over transaction behavior
- **Multiple database support** - PostgreSQL, SQLite, ClickHouse drivers included
- **CLI tool** - Built-in CLI for managing migrations
- **Programmatic API** - Use migrations directly in your Go code
- **Standard library only** - No external dependencies 

## Installation

```bash
go get github.com/alexisvisco/amigo
```

## Quick Start

### 1. Setup your project structure

Create the following structure:

```
yourapp/
├── cmd/
│   └── migrate/
│       └── main.go
├── migrations/
│   └── migrations.go
└── go.mod
```

### 2. Create your migration CLI

Create `cmd/migrate/main.go`:

```go
package main

import (
    "database/sql"
    "log"
    "os"
    
    "github.com/alexisvisco/amigo"
    "yourapp/migrations"
    _ "modernc.org/sqlite"
)

func main() {
    db, err := sql.Open("sqlite", "app.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    config := amigo.DefaultConfiguration
    config.DB = db
    config.Driver = amigo.NewSQLiteDriver("")
    
    // Load migrations
    migrationList := migrations.Migrations(config)
    
    cli := amigo.NewCLI(amigo.CLIConfig{
        Config:               config,
        Migrations:           migrationList,
        Directory:            "migrations",
        DefaultTransactional: true,
        DefaultFileFormat:    "sql",
    })
    
    os.Exit(cli.Run(os.Args[1:]))
}
```

Create `migrations/migrations.go`:

```go
package migrations

import (
    "embed"
    
    "github.com/alexisvisco/amigo"
)

//go:embed *.sql
var sqlFiles embed.FS

func Migrations(cfg amigo.Configuration) []amigo.Migration {
    return []amigo.Migration{}
}
```

**Note**: SQL migrations are embedded using `embed.FS`, making your migration binary portable with no external SQL files needed.

### 3. Generate your first migration

```bash
go run cmd/migrate/main.go generate create_users_table
```

This creates `migrations/20240101120000_create_users_table.sql`:

```sql
-- migrate:up tx=true
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL
);

-- migrate:down tx=true
DROP TABLE users;
```

And automatically updates `migrations/migrations.go`:

```go
package migrations

import (
    "embed"
    
    "github.com/alexisvisco/amigo"
)

//go:embed *.sql
var sqlFiles embed.FS

func Migrations(cfg amigo.Configuration) []amigo.Migration {
    return []amigo.Migration{
        amigo.SQLFileToMigration(sqlFiles, "20240101120000_create_users_table.sql", cfg),
    }
}
```

The `//go:embed *.sql` directive embeds all SQL files into the binary, and `SQLFileToMigration` takes the embedded filesystem as its first argument.

### 4. Run migrations

```bash
# Apply all pending migrations
go run cmd/migrate/main.go up

# View status
go run cmd/migrate/main.go status
```

### 5. (Optional) Build the migration binary

```bash
go build -o bin/migrate cmd/migrate/main.go

# Use it
./bin/migrate up
./bin/migrate status
```

## CLI Commands

### `generate` - Create a new migration

```bash
# Generate SQL migration
go run cmd/migrate/main.go generate create_users_table

# Generate Go migration
go run cmd/migrate/main.go generate --format=go add_email_validation
```

### `up` - Apply pending migrations

```bash
# Apply all pending migrations
go run cmd/migrate/main.go up

# Apply next 2 migrations
go run cmd/migrate/main.go up --steps=2

# Skip confirmation
go run cmd/migrate/main.go up --yes
```

### `down` - Revert applied migrations

```bash
# Revert last migration
go run cmd/migrate/main.go down

# Revert last 3 migrations
go run cmd/migrate/main.go down --steps=3

# Revert all migrations
go run cmd/migrate/main.go down --steps=-1

# Skip confirmation
go run cmd/migrate/main.go down --yes
```

### `status` - Show migration status

```bash
go run cmd/migrate/main.go status
```

Output:
```
Migration Status: 2 applied, 1 pending

Status   Date            Name          Applied At
pending  20240103100000  add_comments
applied  20240102150000  add_posts     2024-01-02 15:30:45
applied  20240101120000  create_users  2024-01-01 12:05:23
```

### `show-config` - Display configuration

```bash
go run cmd/migrate/main.go show-config
```

## Using Migrations Programmatically (Without CLI)

You can run migrations directly in your Go code without using the CLI:

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    
    "github.com/alexisvisco/amigo"
    "yourapp/migrations"
    _ "modernc.org/sqlite"
)

func main() {
    // Open database
    db, err := sql.Open("sqlite", "app.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Configure amigo
    config := amigo.DefaultConfiguration
    config.DB = db
    config.Driver = amigo.NewSQLiteDriver("schema_migrations")
    
    // Load migrations
    migrationList := migrations.Migrations(config)
    
    // Create runner
    runner := amigo.NewRunner(config)
    ctx := context.Background()
    
    // Run all pending migrations
    err = runner.Up(ctx, migrationList)
    if err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }
    
    fmt.Println("Migrations applied successfully!")
}
```

### With Progress Feedback

Use iterators for real-time progress:

```go
// Run migrations with progress feedback
for result := range runner.UpIterator(ctx, migrationList) {
    if result.Error != nil {
        log.Fatalf("Migration failed: %v", result.Error)
    }
    fmt.Printf("✓ %s (%.2fs)\n", result.Migration.Name(), result.Duration.Seconds())
}
```

### Revert Migrations

```go
// Revert last migration
err = runner.Down(ctx, migrationList, amigo.RunnerDownOptionSteps(1))

// Revert with progress
for result := range runner.DownIterator(ctx, migrationList, amigo.RunnerDownOptionSteps(1)) {
    if result.Error != nil {
        log.Fatalf("Revert failed: %v", result.Error)
    }
    fmt.Printf("✓ Reverted %s (%.2fs)\n", result.Migration.Name(), result.Duration.Seconds())
}
```

### Check Migration Status

```go
statuses, err := runner.GetMigrationsStatuses(ctx, migrationList)
if err != nil {
    log.Fatal(err)
}

for _, status := range statuses {
    if status.Applied {
        fmt.Printf("✓ %s (applied at %s)\n", 
            status.Migration.Name, 
            status.Migration.AppliedAt.Format("2006-01-02 15:04:05"))
    } else {
        fmt.Printf("○ %s (pending)\n", status.Migration.Name)
    }
}
```

## Writing Migrations

### SQL Migrations

SQL migrations use annotations to separate up and down migrations:

```sql
-- migrate:up tx=true
CREATE TABLE posts (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    body TEXT
);

CREATE INDEX idx_posts_title ON posts(title);

-- migrate:down tx=true
DROP TABLE posts;
```

#### Transaction Control

Control transaction behavior per migration:

```sql
-- migrate:up tx=false
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);

-- migrate:down tx=false
DROP INDEX CONCURRENTLY idx_users_email;
```

### Go Migrations

Go migrations give you full programmatic control:

```go
package migrations

import (
    "context"
    "database/sql"
    
    "github.com/alexisvisco/amigo"
)

type Migration20240101120000CreateUsers struct{}

func (m Migration20240101120000CreateUsers) Name() string {
    return "create_users"
}

func (m Migration20240101120000CreateUsers) Date() int64 {
    return 20240101120000
}

func (m Migration20240101120000CreateUsers) Up(ctx context.Context, db *sql.DB) error {
    return amigo.Tx(ctx, db, func(tx *sql.Tx) error {
        _, err := tx.Exec(`
            CREATE TABLE users (
                id INTEGER PRIMARY KEY,
                name TEXT NOT NULL
            )
        `)
        return err
    })
}

func (m Migration20240101120000CreateUsers) Down(ctx context.Context, db *sql.DB) error {
    return amigo.Tx(ctx, db, func(tx *sql.Tx) error {
        _, err := tx.Exec(`DROP TABLE users`)
        return err
    })
}
```

#### Without Transactions

```go
func (m Migration20240101120000CreateUsers) Up(ctx context.Context, db *sql.DB) error {
    _, err := db.ExecContext(ctx, `CREATE INDEX CONCURRENTLY idx_users_email ON users(email)`)
    return err
}
```

## Configuration

### Migration Configuration

```go
config := amigo.Configuration{
    DB:                    db,
    Driver:                driver,
    SQLFileUpAnnotation:   "-- migrate:up",
    SQLFileDownAnnotation: "-- migrate:down",
}
```

### CLI Configuration

```go
cliConfig := amigo.CLIConfig{
    Config:               config,
    Migrations:           migrationList,
    Output:               os.Stdout,
    ErrorOut:             os.Stderr,
    Directory:            "db/migrations",
    DefaultTransactional: true,
    DefaultFileFormat:    "sql",
}

cli := amigo.NewCLI(cliConfig)
```

## Database Drivers

### PostgreSQL

```go
import (
    "github.com/alexisvisco/amigo"
    _ "github.com/lib/pq"
)

driver := amigo.NewPostgresDriver("schema_migrations")
```

### SQLite

```go
import (
    "github.com/alexisvisco/amigo"
    _ "modernc.org/sqlite"
)

driver := amigo.NewSQLiteDriver("schema_migrations")
```

### ClickHouse

```go
import (
    "github.com/alexisvisco/amigo"
    _ "github.com/ClickHouse/clickhouse-go/v2"
)

// For standalone ClickHouse
driver := amigo.NewClickHouseDriver("schema_migrations", "")

// For clustered ClickHouse
driver := amigo.NewClickHouseDriver("schema_migrations", "my_cluster")
```

**Note**: When using a cluster, the driver creates a `ReplicatedReplacingMergeTree` table and uses soft deletes for migration rollbacks. For standalone setups (empty cluster string), it uses `MergeTree` and hard deletes.

## Multi-Database Setup

If you have multiple databases (e.g., PostgreSQL for main data and ClickHouse for analytics), create separate migration CLIs:

### PostgreSQL Migration CLI (`cmd/migrate-postgres/main.go`)

```go
package main

import (
    "database/sql"
    "log"
    "os"
    
    "github.com/alexisvisco/amigo"
    "yourapp/migrations/postgres"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/mydb?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    config := amigo.DefaultConfiguration
    config.DB = db
    config.Driver = amigo.NewPostgresDriver("schema_migrations")
    
    migrationList := postgres.Migrations(config)
    
    cli := amigo.NewCLI(amigo.CLIConfig{
        Config:               config,
        Migrations:           migrationList,
        Directory:            "migrations/postgres",
        DefaultTransactional: true,
        DefaultFileFormat:    "sql",
    })
    
    os.Exit(cli.Run(os.Args[1:]))
}
```

### ClickHouse Migration CLI (`cmd/migrate-clickhouse/main.go`)

```go
package main

import (
    "database/sql"
    "log"
    "os"
    
    "github.com/alexisvisco/amigo"
    "yourapp/migrations/clickhouse"
    _ "github.com/ClickHouse/clickhouse-go/v2"
)

func main() {
    db, err := sql.Open("clickhouse", "clickhouse://localhost:9000/default")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    config := amigo.DefaultConfiguration
    config.DB = db
    config.Driver = amigo.NewClickHouseDriver("schema_migrations", "")
    
    migrationList := clickhouse.Migrations(config)
    
    cli := amigo.NewCLI(amigo.CLIConfig{
        Config:               config,
        Migrations:           migrationList,
        Directory:            "migrations/clickhouse",
        DefaultTransactional: true,
        DefaultFileFormat:    "sql",
    })
    
    os.Exit(cli.Run(os.Args[1:]))
}
```

### Directory Structure

```
yourapp/
├── cmd/
│   ├── migrate-postgres/
│   │   └── main.go
│   └── migrate-clickhouse/
│       └── main.go
├── migrations/
│   ├── postgres/
│   │   ├── migrations.go
│   │   ├── 20240101120000_create_users.sql
│   │   └── 20240102150000_create_orders.sql
│   └── clickhouse/
│       ├── migrations.go
│       ├── 20240101120000_create_events.sql
│       └── 20240102150000_create_analytics.sql
└── go.mod
```

### Usage

```bash
# PostgreSQL migrations
go run cmd/migrate-postgres/main.go generate create_users
go run cmd/migrate-postgres/main.go up
go run cmd/migrate-postgres/main.go status

# ClickHouse migrations
go run cmd/migrate-clickhouse/main.go generate create_events
go run cmd/migrate-clickhouse/main.go up
go run cmd/migrate-clickhouse/main.go status
```

### Building Separate Binaries

```bash
# Build both migration tools
go build -o bin/migrate-postgres cmd/migrate-postgres/main.go
go build -o bin/migrate-clickhouse cmd/migrate-clickhouse/main.go

# Use them
./bin/migrate-postgres up
./bin/migrate-clickhouse up
```

## Migration File Format

Migration files follow the format: `{timestamp}_{name}.{ext}`

- Timestamp: `YYYYMMDDHHMMSS`
- Name: Snake case description
- Extension: `.sql` or `.go`

Example: `20240101120000_create_users_table.sql`

## Transaction Helper

Use the `Tx` helper for transactional Go migrations:

```go
err := amigo.Tx(ctx, db, func(tx *sql.Tx) error {
    _, err := tx.Exec("INSERT INTO users (name) VALUES (?)", "Alice")
    if err != nil {
        return err
    }
    
    _, err = tx.Exec("INSERT INTO posts (title) VALUES (?)", "First Post")
    return err
})
```

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR.
