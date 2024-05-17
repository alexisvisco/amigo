package entrypoint

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

func MainPostgres(migrations []schema.Migration) {
	opts := createMainOptions(migrations)
	ok, err := amigo.RunPostgresMigrations(opts)

	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to %s database: %v\n", opts.MigrationDirection, err)
		os.Exit(1)
	}

	if !ok {
		os.Exit(1)
	}
}

func createMainOptions(migrations []schema.Migration) *amigo.RunMigrationOptions {
	dsnFlag := flag.String("dsn", "", "URL connection to the database")
	versionFlag := flag.String("version", "", "Apply or rollback a specific version")
	directionFlag := flag.String("direction", "", "Possibles values are: migrate or rollback")
	jsonFlag := flag.Bool("json", false, "Print the output in JSON")
	silentFlag := flag.Bool("silent", false, "Do not print migrations output")
	timeoutFlag := flag.Duration("timeout", time.Minute*2,
		"Timeout for the migration is the time for the whole migrations to be applied") // not working
	dryRunFlag := flag.Bool("dry-run", false, "Dry run the migration will not apply the migration to the database")
	continueOnErrorFlag := flag.Bool("continue-on-error", false,
		"Continue on error will not rollback the migration if an error occurs")
	schemaVersionTableFlag := flag.String("schema-version-table", "mig_schema_versions",
		"Table name for the schema version")
	showSQLFlag := flag.Bool("sql", false, "Print SQL statements")
	stepsFlag := flag.Int("steps", 1, "Number of steps to rollback")
	debugFlag := flag.Bool("debug", false, "Print debug information")

	// Parse flags
	flag.Parse()

	var out io.Writer = os.Stdout
	if *silentFlag {
		out = io.Discard
	}

	amigo.SetupSlog(*showSQLFlag, *debugFlag, *jsonFlag, out)

	if *debugFlag {
		buf := bytes.NewBuffer(nil)

		tw := tabwriter.NewWriter(buf, 1, 1, 1, ' ', 0)
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(tw, " %s\t%v\n", f.Name, f.Value)
		})
		tw.Flush()

		logger.Debug(events.MessageEvent{Message: fmt.Sprintf("flags: \n%s", buf.String())})
	}

	return &amigo.RunMigrationOptions{
		DSN:                *dsnFlag,
		Version:            versionFlag,
		MigrationDirection: types.MigrationDirection(*directionFlag),
		Migrations:         migrations,
		Timeout:            *timeoutFlag,
		DryRun:             *dryRunFlag,
		ContinueOnError:    *continueOnErrorFlag,
		SchemaVersionTable: schema.TableName(*schemaVersionTableFlag),
		ShowSQL:            *showSQLFlag,
		Steps:              stepsFlag,
	}
}
