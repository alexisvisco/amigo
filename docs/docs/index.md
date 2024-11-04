---
sidebar_position: 1
---

[![Go Report Card](https://goreportcard.com/badge/github.com/alexisvisco/amigo)](https://goreportcard.com/report/github.com/alexisvisco/amigo)
[![GoDoc](https://pkg.go.dev/badge/alexisvisco/amigo)](https://pkg.go.dev/alexisvisco/mig)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/v/release/alexisvisco/amigo.svg)](https://github.com/alexisvisco/amigo/releases)
[![Tests](https://github.com/alexisvisco/amigo/actions/workflows/tests.yml/badge.svg)](https://github.com/alexisvisco/amigo/actions/workflows/tests.yml)

# Introduction

## AMIGO - Migrate SQL with Go language.

A Migration In Golang (AMIGO) is a library that allows you to write migrations in Go language.
It provides you with all the benefits of Go, including type safety, simplicity, and strong tooling support.
AMIGO is designed to be easy to use and integrate into existing projects.

General documentation: [https://amigo.alexisvis.co](https://amigo.alexisvis.co)

## Why another migration library?

First thing, I don't have anything against SQL migrations file library (I support them). I appreciate them but sometimes with SQL files you are limited to do complex migrations that imply your models and business logic.

I just like the way activerecord (rails) migration system and I think it's powerful to combine migration and code.

Some libraries offer Go files migrations but they do not offer APIs to interact with the database schema.

This library offer to you a new way to create migrations in Go with a powerful API to interact with the database schema.

## Features

- **Go Language**: The library allows you to write migrations in Go, making it easy to define schema changes in a programming language you are already familiar with.
- **Type Safety**: Writing migrations in Go provides you with all the language's benefits, including type safety, simplicity, and strong tooling support.
- **Version Control**: Migrations are version controlled.
- **Auto Down Migration**: The library generates down migrations when it's possible.
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
amigo migrate # apply the migration
```

## Example of migration

```go
package migrations

/* ... */

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

Running up and down against this migration : 

```
$ amigo migrate
------> migrating: create_user_table version: 20240524110434
-- create_table(table: users, {columns: id, name, email, created_at, updated_at}, {pk: id})
-- add_index(table: users, name: idx_users_name, columns: [name])
------> version migrated: 20240524090434

$ amigo rollback
------> rollback: create_user_table version: 20240524110434
-- drop_table(table: users)
------> version rolled back: 20240524090434
```

Note that you did not have to write the down migration, the library generates it for you when it's possible.



## Supported databases

- Postgres

## Next supported databases

- SQLite
- MySQL
