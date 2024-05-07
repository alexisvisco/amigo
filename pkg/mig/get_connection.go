package mig

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/alexisvisco/mig/pkg/utils/dblog"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lmittmann/tint"
	sqldblogger "github.com/simukti/sqldb-logger"
	"log/slog"
	"os"
	"strings"
	"time"
)

func GetConnection(dsn string, verbose bool) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("dsn is required, example: postgres://user:password@localhost:5432/dbname?sslmode=disable")
	}

	var db *sql.DB

	if strings.HasPrefix(dsn, "postgres") {
		dbx, err := sql.Open("pgx", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}

		db = dbx
	}

	if verbose && db != nil {
		recorder := dblog.NewLogger(slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			TimeFormat: time.Kitchen,
		})))

		db = sqldblogger.OpenDriver(dsn, db.Driver(), recorder)
	}

	if db != nil {
		return db, nil
	}

	return nil, errors.New("unsupported database")
}
