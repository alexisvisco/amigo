package main

import (
	"database/sql"
	"github.com/alexisvisco/amigo/pkg/entrypoint"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	_ "github.com/mattn/go-sqlite3"
	"os"
	migrations "sqlite/migrations"
)

func main() {
	opts, arg := entrypoint.AmigoContextFromFlags()

	db, err := sql.Open("sqlite3", opts.GetRealDSN())
	if err != nil {
		logger.Error(events.MessageEvent{Message: err.Error()})
		os.Exit(1)
	}

	entrypoint.Main(db, arg, migrations.Migrations, opts)
}
