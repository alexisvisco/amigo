package main

import (
	"database/sql"
	migrations "example/pg/migrations"
	"github.com/alexisvisco/amigo/pkg/entrypoint"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
)

func main() {
	opts, arg := entrypoint.AmigoContextFromFlags()

	db, err := sql.Open("pgx", opts.GetRealDSN())
	if err != nil {
		logger.Error(events.MessageEvent{Message: err.Error()})
		os.Exit(1)
	}

	entrypoint.Main(db, arg, migrations.Migrations, opts)
}
