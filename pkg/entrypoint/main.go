package entrypoint

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoctx"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/colors"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

func Main(db *sql.DB, arg amigo.MainArg, migrations []schema.Migration, ctx *amigoctx.Context) {
	am := amigo.NewAmigo(ctx)
	am.SetupSlog(os.Stdout)

	switch arg {
	case amigo.MainArgMigrate, amigo.MainArgRollback:
		dir := types.MigrationDirectionUp
		if arg == amigo.MainArgRollback {
			dir = types.MigrationDirectionDown
		}
		err := am.RunMigrations(amigo.RunMigrationParams{
			DB:         db,
			Direction:  dir,
			Migrations: migrations,
			LogOutput:  os.Stdout,
		})

		if err != nil {
			logger.Error(events.MessageEvent{Message: err.Error()})
			os.Exit(1)
		}
	case amigo.MainArgSkipMigration:
		err := am.SkipMigrationFile(db)
		if err != nil {
			logger.Error(events.MessageEvent{Message: err.Error()})
			os.Exit(1)
		}
	case amigo.MainArgStatus:
		versions, err := am.GetStatus(db)
		if err != nil {
			logger.Error(events.MessageEvent{Message: err.Error()})
			os.Exit(1)
		}

		hasVersion := func(version string) bool {
			for _, v := range versions {
				if v == version {
					return true
				}
			}
			return false
		}

		// show status of 10 last migrations
		b := &strings.Builder{}
		tw := tabwriter.NewWriter(b, 2, 0, 1, ' ', 0)

		defaultMigrations := sliceArrayOrDefault(migrations, 10)

		for i, m := range defaultMigrations {

			key := fmt.Sprintf("(%s) %s", m.Date().UTC().Format(utils.FormatTime), m.Name())
			value := colors.Red("not applied")
			if hasVersion(m.Date().UTC().Format(utils.FormatTime)) {
				value = colors.Green("applied")
			}

			fmt.Fprintf(tw, "%s\t\t%s", key, value)
			if i != len(defaultMigrations)-1 {
				fmt.Fprintln(tw)
			}
		}

		tw.Flush()
		logger.Info(events.MessageEvent{Message: b.String()})
	}
}

func sliceArrayOrDefault[T any](array []T, x int) []T {
	defaultMigrations := array
	if len(array) >= x {
		defaultMigrations = array[len(array)-x:]
	}
	return defaultMigrations
}

func AmigoContextFromFlags() (*amigoctx.Context, amigo.MainArg) {
	dsnFlag := flag.String("dsn", "", "URL connection to the database")
	jsonFlag := flag.Bool("json", false, "Print the output in JSON")
	showSQLFlag := flag.Bool("sql", false, "Print SQL statements")
	schemaVersionTableFlag := flag.String("schema-version-table", "mig_schema_versions",
		"Table name for the schema version")
	debugFlag := flag.Bool("debug", false, "Print debug information")

	versionFlag := flag.String("version", "", "Apply or rollback a specific version")
	timeoutFlag := flag.Duration("timeout", time.Minute*2,
		"Timeout for the migration is the time for the whole migrations to be applied") // not working
	dryRunFlag := flag.Bool("dry-run", false, "Dry run the migration will not apply the migration to the database")
	continueOnErrorFlag := flag.Bool("continue-on-error", false,
		"Continue on error will not rollback the migration if an error occurs")
	stepsFlag := flag.Int("steps", 1, "Number of steps to rollback")
	showSQLSyntaxHighlightingFlag := flag.Bool("sql-syntax-highlighting", false,
		"Print SQL statements with syntax highlighting")

	// Parse flags
	flag.Parse()

	if flag.NArg() == 0 {
		logger.Error(events.MessageEvent{Message: "missing argument"})
		os.Exit(1)
	}

	arg := amigo.MainArg(flag.Arg(0))
	if err := arg.Validate(); err != nil {
		logger.Error(events.MessageEvent{Message: err.Error()})
		os.Exit(1)
	}

	a := &amigoctx.Context{
		Root: &amigoctx.Root{
			AmigoFolderPath:           "",
			DSN:                       *dsnFlag,
			JSON:                      *jsonFlag,
			ShowSQL:                   *showSQLFlag,
			MigrationFolder:           "",
			PackagePath:               "",
			SchemaVersionTable:        *schemaVersionTableFlag,
			ShellPath:                 "",
			PGDumpPath:                "",
			Debug:                     *debugFlag,
			ShowSQLSyntaxHighlighting: *showSQLSyntaxHighlightingFlag,
		},
	}

	switch arg {
	case amigo.MainArgMigrate:
		a.Migration = &amigoctx.Migration{
			Version:         *versionFlag,
			DryRun:          *dryRunFlag,
			ContinueOnError: *continueOnErrorFlag,
			Timeout:         *timeoutFlag,
		}
	case amigo.MainArgRollback:
		a.Migration = &amigoctx.Migration{
			Version:         *versionFlag,
			ContinueOnError: *continueOnErrorFlag,
			Timeout:         *timeoutFlag,
			Steps:           *stepsFlag,
			DryRun:          *dryRunFlag,
		}
	case amigo.MainArgSkipMigration:
		a.Create = &amigoctx.Create{
			Version: *versionFlag,
		}
	}

	return a, arg
}
