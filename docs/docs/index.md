---
sidebar_position: 1
---

# Introduction

## MIG - Migrate SQL with Go language. 

MIG is a SQL migration tool that uses Go language to write migrations.

## Features

- **Go Language**: The library allows you to write migrations in Go, making it easy to define schema changes in a programming language you are already familiar with.
- **Type Safety**: Writing migrations in Go provides you with all the language's benefits, including type safety, simplicity, and strong tooling support.
- **Version Control**: Migrations are version controlled.
- **Compatibility**: The library supports working with already migrated databases and allows you to seamlessly integrate it into existing projects.

## Installation

To install the library, run the following command:

```sh
go install github.com/alexisvisco/mig
```

## First usage

```sh 
mig context --dsn "postgres://user:password@localhost:5432/dbname" # optional but it avoid to pass the dsn each time
mig init # create the migrations folder, the main file to run migration
mit migrate # apply the migration
```

## Example of migration

```templ
package migrations

import (
    "github.com/alexisvisco/mig/pkg/schema/pg"
    "github.com/alexisvisco/mig/pkg/schema"
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



