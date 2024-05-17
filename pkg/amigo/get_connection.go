package amigo

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/alexisvisco/amigo/pkg/utils/dblog"
	_ "github.com/jackc/pgx/v5/stdlib"
	sqldblogger "github.com/simukti/sqldb-logger"
	"strings"
)

func GetConnection(dsn string) (*sql.DB, *dblog.Logger, error) {
	if dsn == "" {
		return nil, nil, errors.New("dsn is required, example: postgres://user:password@localhost:5432/dbname?sslmode=disable")
	}

	var db *sql.DB

	if strings.HasPrefix(dsn, "postgres") {
		dbx, err := sql.Open("pgx", dsn)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
		}

		db = dbx
	}

	dblogger := dblog.NewLogger()
	db = sqldblogger.OpenDriver(dsn, db.Driver(), dblogger)

	if db != nil {
		return db, dblogger, nil
	}

	return nil, nil, errors.New("unsupported database")
}
