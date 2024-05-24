package pg

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/dblog"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/alexisvisco/amigo/pkg/utils/testutils"
	"github.com/georgysavva/scany/v2/dbscan"
	_ "github.com/jackc/pgx/v5/stdlib"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"strings"
	"testing"
)

var (
	postgresUser = testutils.EnvOrDefault("POSTGRES_USER", "postgres")
	postgresPass = testutils.EnvOrDefault("POSTGRES_PASS", "postgres")
	postgresHost = testutils.EnvOrDefault("POSTGRES_HOST", "localhost")
	postgresPort = testutils.EnvOrDefault("POSTGRES_PORT", "6666")
	postgresDB   = testutils.EnvOrDefault("POSTGRES_DB", "postgres")

	db = schema.DatabaseCredentials{
		User: postgresUser,
		Pass: postgresPass,
		Host: postgresHost,
		Port: postgresPort,
		DB:   postgresDB,
	}

	conn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", postgresUser, postgresPass, postgresHost, postgresPort,
		postgresDB)
)

func connect(t *testing.T) (*sql.DB, dblog.DatabaseLogger) {

	db, err := sql.Open("pgx", conn)
	require.NoError(t, err)

	logger.ShowSQLEvents = true
	slog.SetDefault(slog.New(logger.NewHandler(os.Stdout, &logger.Options{})))
	recorder := dblog.NewHandler(true)

	db = sqldblogger.OpenDriver(conn, db.Driver(), recorder)

	return db, recorder
}

func initSchema(t *testing.T, name string, number ...int32) (*sql.DB, dblog.DatabaseLogger, *schema.Migrator[*Schema], string) {
	conn, recorder := connect(t)
	t.Cleanup(func() {
		_ = conn.Close()
	})

	schemaName := name
	if len(number) > 0 {
		schemaName = fmt.Sprintf("%s_%d", name, number[0])
	}

	_, err := conn.ExecContext(context.Background(), fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	require.NoError(t, err)
	_, err = conn.ExecContext(context.Background(), fmt.Sprintf("CREATE SCHEMA %s", schemaName))
	require.NoError(t, err)

	mig := schema.NewMigrator(context.Background(), conn, NewPostgres, &schema.MigratorOption{})

	return conn, recorder, mig, schemaName
}

func TestPostgres_AddExtension(t *testing.T) {
	sc := "tst_pg_add_extension"

	// this also test the DropExtension method

	// because of the weirdness of the extensions system, extensions are not relative to a schema, but are INSTALLED
	// in a schema.
	// If you add an extension in schema1 then in schema 2 you will have an error because the extension is already installed ...
	// So tests cannot be run in parallel.

	t.Run("with schema", func(t *testing.T) {
		p, r, schemaName := baseTest(t, "select 1", sc)

		p.DropExtension("hstore", schema.DropExtensionOptions{IfExists: true})
		p.AddExtension("hstore", schema.ExtensionOptions{Schema: schemaName})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("without schema", func(t *testing.T) {
		p, r, _ := baseTest(t, "select 1", sc, 1)

		p.DropExtension("hstore", schema.DropExtensionOptions{IfExists: true})
		p.AddExtension("hstore", schema.ExtensionOptions{})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("with IfNotExists", func(t *testing.T) {
		p, _, schemaName := baseTest(t, "select 1", sc, 2)

		p.DropExtension("hstore", schema.DropExtensionOptions{IfExists: true})
		p.AddExtension("hstore", schema.ExtensionOptions{Schema: schemaName})

		require.Panics(t, func() {
			p.AddExtension("hstore", schema.ExtensionOptions{Schema: schemaName})
		})

		t.Run("ensure no panic if param IfNotExists is true", func(t *testing.T) {
			p.AddExtension("hstore", schema.ExtensionOptions{IfNotExists: true, Schema: schemaName})
		})
	})
}

func assertConstraintExist(t *testing.T, p *Schema, tableName schema.TableName, constraintName string) {
	require.True(t, p.ConstraintExist(tableName, constraintName))
}

func assertConstraintNotExist(t *testing.T, p *Schema, tableName schema.TableName, constraintName string) {
	require.False(t, p.ConstraintExist(tableName, constraintName))
}

func asserIndexExist(t *testing.T, p *Schema, tableName schema.TableName, indexName string) {
	require.True(t, p.IndexExist(tableName, indexName))
}

func baseTest(t *testing.T, init string, schema string, number ...int32) (postgres *Schema, rec dblog.DatabaseLogger, schem string) {
	conn, rec, mig, schem := initSchema(t, schema, number...)

	replacer := utils.Replacer{
		"schema": utils.StrFunc(schem),
	}

	if init != "" {
		_, err := conn.ExecContext(context.Background(), replacer.Replace(init))
		require.NoError(t, err)
	}

	p := mig.NewSchema()

	rec.ToggleLogger(true)
	rec.SetRecord(true)

	return p, rec, schem
}

type columnInfo struct {
	ColumnName    string
	DataType      string
	ColumnDefault *string
	PrimaryKey    bool
}

func dumpColumns(t *testing.T, p *Schema, tableName schema.TableName) []columnInfo {
	var columns []columnInfo

	query := `select c.column_name,
       c.data_type,
       c.column_default,
       case when tc.constraint_type = 'PRIMARY KEY' then true else false end as primary_key
from information_schema.Columns c
         LEFT JOIN information_schema.key_column_usage kcu
                   on c.table_name = kcu.table_name and c.column_name = kcu.column_name and c.table_schema = kcu.table_schema
         LEFT JOIN information_schema.table_constraints tc
                   on kcu.constraint_name = tc.constraint_name and kcu.table_name = tc.table_name and
                      kcu.table_schema = tc.table_schema and tc.constraint_type = 'PRIMARY KEY'

where c.table_schema = $1
  and c.table_name = $2
order by column_name;`

	rows, err := p.DB.QueryContext(context.Background(), query, tableName.Schema(), tableName.Name())
	require.NoError(t, err)

	require.NoError(t, dbscan.ScanAll(&columns, rows))

	for i := range columns {
		if columns[i].ColumnDefault != nil && strings.Contains(*columns[i].ColumnDefault, "nextval") {
			columns[i].ColumnDefault = utils.Ptr("nextval")
		}
	}

	return columns
}
