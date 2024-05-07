package entrypoint

import (
	"flag"
	"fmt"
	"github.com/alexisvisco/mig/pkg/mig"
	"github.com/alexisvisco/mig/pkg/schema"
	"github.com/alexisvisco/mig/pkg/types"
	"github.com/alexisvisco/mig/pkg/utils/tracker"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

func MainPostgres(migrations []schema.Migration) {
	opts := createMainOptions(migrations)
	ok, err := mig.MigratePostgres(opts)

	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to %s database: %v\n", opts.MigrationDirection, err)
		os.Exit(1)
	}

	if !ok {
		os.Exit(1)
	}
}

func createMainOptions(migrations []schema.Migration) *mig.MainOptions {
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
	schemaVersionTableFlag := flag.String("schema-version-table", "schema_version", "Table name for the schema version")
	verboseFlag := flag.Bool("verbose", false, "Print SQL statements")
	stepsFlag := flag.Int("steps", 1, "Number of steps to rollback")

	// Parse flags
	flag.Parse()

	if *verboseFlag {
		fmt.Println("-- flags:")
		tw := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(tw, " %s\t%v\n", f.Name, f.Value)
		})
		tw.Flush()
	}

	var out io.Writer = os.Stdout
	if *silentFlag {
		out = io.Discard
	}

	t := tracker.NewLogger(*jsonFlag, out)

	return &mig.MainOptions{
		DSN:                *dsnFlag,
		Version:            versionFlag,
		MigrationDirection: types.MigrationDirection(*directionFlag),
		Migrations:         migrations,
		Timeout:            *timeoutFlag,
		DryRun:             *dryRunFlag,
		ContinueOnError:    *continueOnErrorFlag,
		SchemaVersionTable: schema.TableName(*schemaVersionTableFlag),
		Verbose:            *verboseFlag,
		Steps:              stepsFlag,
		Tracker:            t,
	}
}
