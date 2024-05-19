package main

import (
	"database/sql"
	"fmt"
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/testutils"
	migrationswithchange "github.com/alexisvisco/amigo/testdata/e2e/pg/migrations_with_change"
	migrationwithclassic "github.com/alexisvisco/amigo/testdata/e2e/pg/migrations_with_classic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path"
	"testing"
	"time"
)

var (
	greenColor = "\033[32m"
	resetColor = "\033[0m"
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

	connection, _, _ := amigo.GetConnection(conn)
	defer connection.Close()

	t.Run("migration_with_change", func(t *testing.T) {
		s := "migrations_with_change"

		base := amigo.RunMigrationOptions{
			DSN:                conn,
			SchemaVersionTable: "migrations_with_change.mig_schema_versions",
			Timeout:            time.Minute * 2,
			Migrations:         migrationswithchange.Migrations,
			//ShowSQL:            true,
		}

		createSchema(t, connection, s)

		ensureMigrationsAreReversible(t, base, db, s)
	})

	t.Run("migration_with_classic", func(t *testing.T) {
		s := "migrations_with_classic"

		base := amigo.RunMigrationOptions{
			DSN:                conn,
			SchemaVersionTable: "migrations_with_classic.mig_schema_versions",
			Timeout:            time.Minute * 2,
			Migrations:         migrationwithclassic.Migrations,
			//ShowSQL:            true,
		}

		createSchema(t, connection, s)

		ensureMigrationsAreReversible(t, base, db, s)
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
func ensureMigrationsAreReversible(t *testing.T, base amigo.RunMigrationOptions, db schema.DatabaseCredentials, s string) {
	for i := 0; i < len(migrationswithchange.Migrations); i++ {

		fmt.Println()
		fmt.Println(greenColor, "Running migration", migrationswithchange.Migrations[i].Name(), resetColor)
		fmt.Println()

		version := migrationswithchange.Migrations[i].Date().UTC().Format(utils.FormatTime)
		id := path.Join(s, fmt.Sprintf("%s_%s", version, migrationswithchange.Migrations[i].Name()))

		ok, err := amigo.RunPostgresMigrations(mergeOptions(base, amigo.RunMigrationOptions{
			MigrationDirection: types.MigrationDirectionUp,
			Version:            utils.Ptr(version),
		}))

		assert.True(t, ok, "migration %s failed", migrationswithchange.Migrations[i].Name())
		assert.NoError(t, err, "migration %s failed", migrationswithchange.Migrations[i].Name())

		testutils.MaySnapshotSavePgDump(t, s, db, id, false)

		if i == 0 {
			continue
		}

		ok, err = amigo.RunPostgresMigrations(mergeOptions(base, amigo.RunMigrationOptions{
			MigrationDirection: types.MigrationDirectionDown,
		}))

		assert.True(t, ok, "rollback migration %s failed", migrationswithchange.Migrations[i].Name())
		assert.NoError(t, err, "rollback migration %s failed", migrationswithchange.Migrations[i].Name())

		oldVersion := migrationswithchange.Migrations[i-1].Date().UTC().Format(utils.FormatTime)
		oldId := path.Join(s, fmt.Sprintf("%s_%s", oldVersion, migrationswithchange.Migrations[i-1].Name()))

		testutils.AssertSnapshotPgDumpDiff(t, s, db, oldId)

		// here we have verified that the rollback is correct

		ok, err = amigo.RunPostgresMigrations(mergeOptions(base, amigo.RunMigrationOptions{
			MigrationDirection: types.MigrationDirectionUp,
			Version:            utils.Ptr(migrationswithchange.Migrations[i].Date().UTC().Format(utils.FormatTime)),
		}))

		assert.True(t, ok, "migration %s failed", migrationswithchange.Migrations[i].Name())
		assert.NoError(t, err, "migration %s failed", migrationswithchange.Migrations[i].Name())

		testutils.AssertSnapshotPgDumpDiff(t, s, db, id)

		// here we have verified that the up is correct with the previous rollback
	}

	fmt.Println()
	fmt.Println(greenColor, "Rollback all migrations", resetColor)
	fmt.Println()

	ok, err := amigo.RunPostgresMigrations(mergeOptions(base, amigo.RunMigrationOptions{
		MigrationDirection: types.MigrationDirectionDown,
		Steps:              utils.Ptr(len(migrationswithchange.Migrations) - 1),
	}))

	assert.True(t, ok, "rollback all migrations failed")
	assert.NoError(t, err, "rollback all migrations failed")

	firstVersion := migrationswithchange.Migrations[0].Date().UTC().Format(utils.FormatTime)
	firstId := path.Join(s, fmt.Sprintf("%s_%s", firstVersion, migrationswithchange.Migrations[0].Name()))
	testutils.AssertSnapshotPgDumpDiff(t, s, db, firstId)

	fmt.Println()
	fmt.Println(greenColor, "Up all migrations", resetColor)
	fmt.Println()

	ok, err = amigo.RunPostgresMigrations(mergeOptions(base, amigo.RunMigrationOptions{
		MigrationDirection: types.MigrationDirectionUp,
	}))

	assert.True(t, ok, "up all migrations failed")
	assert.NoError(t, err, "up all migrations failed")

	lastVersion := migrationswithchange.Migrations[len(migrationswithchange.Migrations)-1].Date().UTC().Format(utils.FormatTime)
	lastId := path.Join(s, fmt.Sprintf("%s_%s", lastVersion,
		migrationswithchange.Migrations[len(migrationswithchange.Migrations)-1].Name()))

	testutils.AssertSnapshotPgDumpDiff(t, s, db, lastId)

	// here we have verified that the up all is correct
}

func mergeOptions(b amigo.RunMigrationOptions, options ...amigo.RunMigrationOptions) *amigo.RunMigrationOptions {
	base := amigo.RunMigrationOptions{
		DSN:                b.DSN,
		MigrationDirection: b.MigrationDirection,
		Version:            b.Version,
		Steps:              b.Steps,
		SchemaVersionTable: b.SchemaVersionTable,
		DryRun:             b.DryRun,
		ContinueOnError:    b.ContinueOnError,
		Timeout:            b.Timeout,
		Migrations:         b.Migrations,
		JSON:               b.JSON,
		ShowSQL:            b.ShowSQL,
		Debug:              b.Debug,
		Shell:              b.Shell,
	}

	for _, opt := range options {
		if opt.DSN != "" {
			base.DSN = opt.DSN
		}

		if opt.SchemaVersionTable != "" {
			base.SchemaVersionTable = opt.SchemaVersionTable
		}

		if opt.Timeout != 0 {
			base.Timeout = opt.Timeout
		}

		if opt.Migrations != nil {
			base.Migrations = opt.Migrations
		}

		if opt.ShowSQL {
			base.ShowSQL = opt.ShowSQL
		}

		if opt.Debug {
			base.Debug = opt.Debug
		}

		if opt.JSON {
			base.JSON = opt.JSON
		}

		if opt.DryRun {
			base.DryRun = opt.DryRun
		}

		if opt.ContinueOnError {
			base.ContinueOnError = opt.ContinueOnError
		}

		if opt.Version != nil {
			base.Version = opt.Version
		}

		if opt.Steps != nil {
			base.Steps = opt.Steps
		}

		if opt.Shell != "" {
			base.Shell = opt.Shell
		}

		if opt.MigrationDirection != "" {
			base.MigrationDirection = opt.MigrationDirection
		}
	}

	return &base
}
