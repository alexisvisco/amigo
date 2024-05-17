[![GoDoc](https://pkg.go.dev/badge/alexisvisco/mig)](https://pkg.go.dev/alexisvisco/mig)

# Introduction

## MIG - Migrate SQL with Go language.

Migration In Golang (MIG) is a library that allows you to write migrations in Go language. 
It provides you with all the benefits of Go, including type safety, simplicity, and strong tooling support.
MIG is designed to be easy to use and integrate into existing projects.

General documentation: [https://mig-go.alexisvis.co](https://mig-go.alexisvis.co)

## Features

- **Go Language**: The library allows you to write migrations in Go, making it easy to define schema changes in a programming language you are already familiar with.
- **Type Safety**: Writing migrations in Go provides you with all the language's benefits, including type safety, simplicity, and strong tooling support.
- **Version Control**: Migrations are version controlled.
- **Compatibility**: The library supports working with already migrated databases and allows you to seamlessly integrate it into existing projects.

## Installation

To install the library, run the following command:

```sh
go install github.com/alexisvisco/amigo@latest
```

## First usage

```sh 
amigo context --dsn "postgres://user:password@localhost:5432/dbname" # optional but it avoid to pass the dsn each time
amigo init # create the migrations folder, the main file to run migration
mit migrate # apply the migration
```

## Example of migration

```go
package migrations

import (
    "github.com/alexisvisco/amigo/pkg/schema/pg"
    "github.com/alexisvisco/amigo/pkg/schema"
    "time"
)

type Migration20240502155033SchemaVersion struct {}

func (m Migration20240502155033SchemaVersion) Change(s *pg.Schema) {
    s.CreateTable("public.mig_schema_versions", func(s *pg.PostgresTableDef) {
        s.String("id")
    })
}

func (m Migration20240502155033SchemaVersion) Name() string {
    return "schema_version"
}

func (m Migration20240502155033SchemaVersion) Date() time.Time {
    t, _  := time.Parse(time.RFC3339, "2024-05-02T17:50:33+02:00")
    return t
}
```


## Supported databases 

- Postgres

## Next supported databases

- SQLite
- MySQL
