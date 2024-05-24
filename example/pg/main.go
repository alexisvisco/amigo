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
	db, err := sql.Open("pgx", "")
	if err != nil {
		panic(err)
	}

	err = amigo.NewAmigo(amigoctx.MergeContext(amigoctx.Context{
		Root: &amigoctx.Root{
			DSN:     dsn,
			ShowSQL: true,
			JSON:    true,
		},
	})).RunMigrations(amigo.RunMigrationParams{
		DB:         db,
		Direction:  types.MigrationDirectionUp,
		Migrations: migrations.Migrations,
		LogOutput:  os.Stdout,
	})
	if err != nil {
		panic(err)
	}
}
