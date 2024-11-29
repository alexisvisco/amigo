package main

import (
	"database/sql"
	"example/pg/migrations"
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

	err = amigo.NewAmigo(amigoconfig.NewContext().WithDSN(dsn)).RunMigrations(amigo.RunMigrationParams{
		DB:         db,
		Direction:  types.MigrationDirectionDown,
		Migrations: migrations.Migrations,
		LogOutput:  os.Stdout,
	})
	if err != nil {
		panic(err)
	}
}
