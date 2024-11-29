package pg

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/dblog"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/alexisvisco/amigo/pkg/utils/testutils"
	_ "github.com/jackc/pgx/v5/stdlib"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/stretchr/testify/require"
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

func versionTable(schemaName string, s *Schema) {
	s.CreateTable(schema.Table("mig_schema_version", schemaName), func(s *PostgresTableDef) {
		s.String("version")
	}, schema.TableOptions{IfNotExists: true})
}

func connect(t *testing.T) (*sql.DB, dblog.DatabaseLogger) {

	db, err := sql.Open("pgx", conn)
	require.NoError(t, err)

	logger.ShowSQLEvents = true
	slog.SetDefault(slog.New(logger.NewHandler(os.Stdout, &logger.Options{})))
	recorder := dblog.NewHandler(true)

	db = sqldblogger.OpenDriver(conn, db.Driver(), recorder)

	return db, recorder
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

	mig := schema.NewMigrator(context.Background(), conn, NewPostgres, amigoconfig.NewConfig())

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

		p.DropExtension("hstore", schema.DropExtensionOptions{IfExists: true, Cascade: true})
		p.AddExtension("hstore", schema.ExtensionOptions{Schema: schemaName})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("without schema", func(t *testing.T) {
		p, r, _ := baseTest(t, "select 1", sc, 1)

		p.DropExtension("hstore", schema.DropExtensionOptions{IfExists: true, Cascade: true})
		p.AddExtension("hstore", schema.ExtensionOptions{})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("with IfNotExists", func(t *testing.T) {
		p, _, schemaName := baseTest(t, "select 1", sc, 2)

		p.DropExtension("hstore", schema.DropExtensionOptions{IfExists: true, Cascade: true})
		p.AddExtension("hstore", schema.ExtensionOptions{Schema: schemaName})

		require.Panics(t, func() {
			p.AddExtension("hstore", schema.ExtensionOptions{Schema: schemaName})
		})

		t.Run("ensure no panic if param IfNotExists is true", func(t *testing.T) {
			p.AddExtension("hstore", schema.ExtensionOptions{IfNotExists: true, Schema: schemaName})
		})
	})
}

func TestPostgres_Versions(t *testing.T) {
	p, _, _ := baseTest(t, "select 1", "tst_pg_add_version")

	versionTable("tst_pg_add_version", p)

	p.Context.Config.SchemaVersionTable = schema.Table("mig_schema_version", "tst_pg_add_version").String()

	p.AddVersion("v1")
	versions := p.FindAppliedVersions()

	require.Len(t, versions, 1)
	require.Equal(t, "v1", versions[0])

	p.RemoveVersion("v1")

	versions = p.FindAppliedVersions()
	require.Len(t, versions, 0)
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

	rows, err := p.TX.QueryContext(context.Background(), query, tableName.Schema(), tableName.Name())
	require.NoError(t, err)

	defer rows.Close() // Ensure rows are closed

	for rows.Next() {
		var col columnInfo
		err := rows.Scan(&col.ColumnName, &col.DataType, &col.ColumnDefault, &col.PrimaryKey)
		require.NoError(t, err)
		columns = append(columns, col)
	}

	err = rows.Err()
	require.NoError(t, err)

	for i := range columns {
		if columns[i].ColumnDefault != nil && strings.Contains(*columns[i].ColumnDefault, "nextval") {
			columns[i].ColumnDefault = utils.Ptr("nextval")
		}
	}

	return columns
}

func TestPostgres_Query(t *testing.T) {
	t.Run("basic query execution", func(t *testing.T) {
		p, _, schemaName := baseTest(t, "", "tst_pg_query")

		// Create a test table
		p.CreateTable(schema.Table("test_query", schemaName), func(s *PostgresTableDef) {
			s.Integer("id")
			s.String("name")
		})

		// Insert some test data
		_, err := p.TX.ExecContext(context.Background(), fmt.Sprintf(
			"INSERT INTO %s.test_query (id, name) VALUES ($1, $2), ($3, $4)",
			schemaName,
		), 1, "Alice", 2, "Bob")
		require.NoError(t, err)

		// Test Query function
		var results []struct {
			ID   int
			Name string
		}

		p.Query(
			fmt.Sprintf("SELECT id, name FROM %s.test_query ORDER BY id", schemaName),
			[]interface{}{},
			func(rows *sql.Rows) error {
				var result struct {
					ID   int
					Name string
				}
				if err := rows.Scan(&result.ID, &result.Name); err != nil {
					return err
				}
				results = append(results, result)
				return nil
			},
		)

		// Verify results
		require.Len(t, results, 2)
		require.Equal(t, 1, results[0].ID)
		require.Equal(t, "Alice", results[0].Name)
		require.Equal(t, 2, results[1].ID)
		require.Equal(t, "Bob", results[1].Name)
	})

	t.Run("query with arguments", func(t *testing.T) {
		p, _, schemaName := baseTest(t, "", "tst_pg_query_args")

		// Create a test table
		p.CreateTable(schema.Table("test_query", schemaName), func(s *PostgresTableDef) {
			s.Integer("id")
			s.String("name")
		})

		// Insert test data
		_, err := p.TX.ExecContext(context.Background(), fmt.Sprintf(
			"INSERT INTO %s.test_query (id, name) VALUES ($1, $2), ($3, $4)",
			schemaName,
		), 1, "Alice", 2, "Bob")
		require.NoError(t, err)

		// Test Query with arguments
		var result string
		p.Query(
			fmt.Sprintf("SELECT name FROM %s.test_query WHERE id = $1", schemaName),
			[]interface{}{1},
			func(rows *sql.Rows) error {
				return rows.Scan(&result)
			},
		)

		require.Equal(t, "Alice", result)
	})

	t.Run("query with error handling", func(t *testing.T) {
		p, _, schemaName := baseTest(t, "", "tst_pg_query_error")

		// Test invalid query
		require.Panics(t, func() {
			p.Query(
				fmt.Sprintf("SELECT * FROM %s.nonexistent_table", schemaName),
				[]interface{}{},
				func(rows *sql.Rows) error {
					return nil
				},
			)
		})

		// Test error in row handler
		p.CreateTable(schema.Table("test_query", schemaName), func(s *PostgresTableDef) {
			s.Integer("id")
		})

		_, err := p.TX.ExecContext(context.Background(), fmt.Sprintf(
			"INSERT INTO %s.test_query (id) VALUES ($1)",
			schemaName,
		), 1)
		require.NoError(t, err)

		require.Panics(t, func() {
			p.Query(
				fmt.Sprintf("SELECT id FROM %s.test_query", schemaName),
				[]interface{}{},
				func(rows *sql.Rows) error {
					return fmt.Errorf("test error in row handler")
				},
			)
		})
	})
}
