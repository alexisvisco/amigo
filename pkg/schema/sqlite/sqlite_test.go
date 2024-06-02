package sqlite

import (
	"context"
	"database/sql"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils/dblog"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	_ "github.com/mattn/go-sqlite3"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"path"
	"testing"
)

func connect(t *testing.T) (*sql.DB, dblog.DatabaseLogger) {
	dbFile := path.Join("testdata", t.Name()) + ".data"
	err := os.MkdirAll(path.Dir(dbFile), 0755)
	require.NoError(t, err)

	err = os.Remove(dbFile)
	if err != nil && !os.IsNotExist(err) {
		require.NoError(t, err)
	}

	conn, err := sql.Open("sqlite3", dbFile)
	require.NoError(t, err)

	logger.ShowSQLEvents = true
	slog.SetDefault(slog.New(logger.NewHandler(os.Stdout, &logger.Options{})))
	recorder := dblog.NewHandler(true)

	conn = sqldblogger.OpenDriver(dbFile, conn.Driver(), recorder)

	return conn, recorder
}

func baseTest(t *testing.T, init string) (postgres *Schema, rec dblog.DatabaseLogger) {
	db, rec := connect(t)

	m := schema.NewMigrator(context.Background(), db, NewSQLite, &schema.MigratorOption{})

	if init != "" {
		_, err := db.ExecContext(context.Background(), init)
		require.NoError(t, err)
	}

	rec.ToggleLogger(true)
	rec.SetRecord(true)

	return m.NewSchema(), rec
}

func asserIndexExist(t *testing.T, p *Schema, tableName schema.TableName, indexName string) {
	require.True(t, p.IndexExist(tableName, indexName))
}
