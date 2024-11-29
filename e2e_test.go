package main

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/colors"
	"github.com/alexisvisco/amigo/pkg/utils/testutils"
	migrationswithchange "github.com/alexisvisco/amigo/testdata/e2e/pg/migrations_with_change"
	migrationswithclassic "github.com/alexisvisco/amigo/testdata/e2e/pg/migrations_with_classic"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_2e2_postgres(t *testing.T) {
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

	connection, err := sql.Open("pgx", conn)
	require.NoError(t, err)

	t.Run("migration_with_change", func(t *testing.T) {
		s := "migrations_with_change"
		createSchema(t, connection, s)
		ensureMigrationsAreReversible(t, db, migrationswithchange.Migrations, connection, conn, s)
	})

	t.Run("migration_with_classic", func(t *testing.T) {
		s := "migrations_with_classic"
		createSchema(t, connection, s)
		ensureMigrationsAreReversible(t, db, migrationswithclassic.Migrations, connection, conn, s)
	})

}

func createSchema(t *testing.T, connection *sql.DB, s string) {
	_, err := connection.Exec("DROP SCHEMA IF EXISTS " + s + " CASCADE")
	require.NoError(t, err)
	_, err = connection.Exec("CREATE SCHEMA " + s)
	require.NoError(t, err)
}

// ensureMigrationsAreReversible ensures that all migrations are reversible and works nice.
// for each migrations :
// up then do a snapshot
// rollback if it's not the first migration, check with the previous snapshot
// up again and compare the snapshot with the captured snapshot
// this until the last migration
//
// then rollback all migrations, check the snapshot with the first one
// then up all migrations, check the snapshot with the last one
func ensureMigrationsAreReversible(t *testing.T, db schema.DatabaseCredentials, migrations []schema.Migration, sql *sql.DB, dsn, schema string) {
	actx := amigoconfig.NewConfig()
	actx.ShowSQL = true
	actx.DSN = dsn
	actx.SchemaVersionTable = schema + ".mig_schema_versions"

	am := amigo.NewAmigo(actx)

	runParamsUp := amigo.RunMigrationParams{
		DB:         sql,
		Direction:  types.MigrationDirectionUp,
		Migrations: migrations,
		LogOutput:  os.Stdout,
	}

	runParamsDown := amigo.RunMigrationParams{
		DB:         sql,
		Direction:  types.MigrationDirectionDown,
		Migrations: migrations,
		LogOutput:  os.Stdout,
	}

	for i := 0; i < len(migrations); i++ {

		printStep(fmt.Sprintf("Step %d/%d", i+1, len(migrations)))

		printStep(fmt.Sprintf("Up migration %s", migrations[i].Name()))

		version := migrations[i].Date().UTC().Format(utils.FormatTime)
		id := path.Join(schema, fmt.Sprintf("%s_%s", version, migrations[i].Name()))

		actx.Migration.Version = version
		err := am.RunMigrations(runParamsUp)
		assert.NoError(t, err, "migration %s failed", migrations[i].Name())

		testutils.MaySnapshotSavePgDump(t, schema, db, id, false)

		if i == 0 {
			continue
		}

		printStep(fmt.Sprintf("Rollback migration %s", migrations[i].Name()))

		actx.Migration.Version = ""
		err = am.RunMigrations(runParamsDown)
		assert.NoError(t, err, "rollback migration %s failed", migrations[i].Name())

		oldVersion := migrations[i-1].Date().UTC().Format(utils.FormatTime)
		oldId := path.Join(schema, fmt.Sprintf("%s_%s", oldVersion, migrations[i-1].Name()))

		testutils.AssertSnapshotPgDumpDiff(t, schema, db, oldId)

		// here we have verified that the rollback is correct

		printStep(fmt.Sprintf("Up migration %s", migrations[i].Name()))

		actx.Migration.Version = version
		err = am.RunMigrations(runParamsUp)
		assert.NoError(t, err, "migration %s failed", migrations[i].Name())

		testutils.AssertSnapshotPgDumpDiff(t, schema, db, id)

		// here we have verified that the up is correct with the previous rollback
	}

	printStep("--------------------")

	printStep("Rollback all migrations")
	actx.Migration.Version = ""
	actx.Migration.Steps = len(migrations) - 1
	err := am.RunMigrations(runParamsDown)
	assert.NoError(t, err, "rollback all migrations failed")

	firstVersion := migrations[0].Date().UTC().Format(utils.FormatTime)
	firstId := path.Join(schema, fmt.Sprintf("%s_%s", firstVersion, migrations[0].Name()))
	testutils.AssertSnapshotPgDumpDiff(t, schema, db, firstId)

	printStep("Up all migrations")

	err = am.RunMigrations(runParamsUp)
	assert.NoError(t, err, "up all migrations failed")

	lastVersion := migrations[len(migrations)-1].Date().UTC().Format(utils.FormatTime)
	lastId := path.Join(schema, fmt.Sprintf("%s_%s", lastVersion,
		migrations[len(migrations)-1].Name()))

	testutils.AssertSnapshotPgDumpDiff(t, schema, db, lastId)

	// here we have verified that the up all is correct
}

func printStep(step string) {
	fmt.Println()
	fmt.Println(colors.Yellow(step))
	fmt.Println()
}
