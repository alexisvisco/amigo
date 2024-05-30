package main

import (
	"database/sql"
	"github.com/alexisvisco/amigo/pkg/entrypoint"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	_ "github.com/go-sql-driver/mysql"
	migrations "mysql/migrations"
	"os"
)

func main() {
	opts, arg := entrypoint.AmigoContextFromFlags()

	db, err := sql.Open("mysql", opts.DSN)
	if err != nil {
		logger.Error(events.MessageEvent{Message: err.Error()})
		os.Exit(1)
	}

	entrypoint.Main(db, arg, migrations.Migrations, opts)
}
