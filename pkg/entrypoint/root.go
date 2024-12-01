package entrypoint

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/spf13/cobra"
)

var (
	config         = amigoconfig.NewConfig()
	provider       func(cfg amigoconfig.Config) (*sql.DB, []schema.Migration, error)
	customAmigoFn  func(a *amigo.Amigo) *amigo.Amigo
	migrationsFile = "migrations.go"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "amigo",
	Short: "Tool to manage database migrations with go files",
	Long: `Basic usage: 
First you need to create a main folder with amigo init:
	
	will create a folder in db with a context file inside to not have to pass the dsn every time.
	
	Postgres:
	$ amigo context --dsn "postgres://user:password@host:port/dbname?sslmode=disable"
	
	SQLite:
	$ amigo context --dsn "sqlite:/path/to/db.sqlite" --schema-version-table mig_schema_versions
	
	Unknown Driver (Mysql in this case):
	$ amigo context --dsn "user:password@tcp(host:port)/dbname"
	
	
	$ amigo init
	note: will create:
	- folder named migrations with a file named migrations.go that contains the list of migrations
	- a new migration to create the  schema version table
	- a main.go in the $amigo_folder
	
Apply migrations:
	$ amigo migrate
	note: you can set --version <version> to migrate a specific version

Create a new migration:
	$ amigo create "create_table_users"
	note: you can set --dump if you already have a database and you want to create the first migration with what's 
	already in the database. --skip will add the version of the created migration inside the schema version table.

Rollback a migration:
	$ amigo rollback
	note: you can set --step <number> to rollback a specific number of migrations, and --version <version> to rollback 
	to a specific version
`,
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&config.AmigoFolderPath, "amigo-folder", "m",
		amigoconfig.DefaultAmigoFolder,
		"Folder path to use for creating amigo related files related to this repository")

	rootCmd.PersistentFlags().StringVar(&config.DSN, "dsn", "",
		"The database connection string example: postgres://user:password@host:port/dbname?sslmode=disable")

	rootCmd.PersistentFlags().BoolVarP(&config.JSON, "json", "j", false, "Output in json format")

	rootCmd.PersistentFlags().StringVar(&config.MigrationFolder, "folder", amigoconfig.DefaultMigrationFolder,
		"Folder where the migrations are stored")

	rootCmd.PersistentFlags().StringVarP(&config.MigrationPackageName, "package", "p",
		amigoconfig.DefaultMigrationPackageName,
		"Package name of the migrations folder")

	rootCmd.PersistentFlags().StringVarP(&config.SchemaVersionTable, "schema-version-table", "t",
		amigoconfig.DefaultSchemaVersionTable, "Table name to keep track of the migrations")

	rootCmd.PersistentFlags().StringVar(&config.ShellPath, "shell-path", amigoconfig.DefaultShellPath,
		"Shell to use (for: amigo create --dump, it uses pg dump command)")

	rootCmd.PersistentFlags().BoolVar(&config.ShowSQL, "sql", false, "Print SQL queries")

	rootCmd.PersistentFlags().BoolVar(&config.ShowSQLSyntaxHighlighting, "sql-syntax-highlighting", true,
		"Print SQL queries with syntax highlighting")

	rootCmd.PersistentFlags().StringVar(&config.SchemaOutPath, "schema-out-path", amigoconfig.DefaultSchemaOutPath,
		"File path of the schema dump if any")

	rootCmd.PersistentFlags().StringVar(&config.PGDumpPath, "pg-dump-path", amigoconfig.DefaultPGDumpPath,
		"Path to the pg_dump command if --dump is set")

	rootCmd.PersistentFlags().StringVar(&config.SchemaToDump, "schema-to-dump",
		amigoconfig.DefaultSchemaToDump, "Schema to use when dumping schema")

	rootCmd.PersistentFlags().BoolVar(&config.Debug, "debug", false, "Print debug information")

	initConfig()
}

func initConfig() {
	yamlConfig, err := amigoconfig.LoadYamlConfig(filepath.Join(config.AmigoFolderPath, amigoconfig.FileName))
	if err != nil {
		logger.Error(events.MessageEvent{Message: fmt.Sprintf("error: can't read config: %s", err)})
		return
	}

	config.OverrideWithYamlConfig(yamlConfig)
}

func wrapCobraFunc(f func(cmd *cobra.Command, am amigo.Amigo, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {

		am := amigo.NewAmigo(config)
		if customAmigoFn != nil {
			am = *customAmigoFn(&am)
		}
		am.SetupSlog(os.Stdout, nil)

		if err := f(cmd, am, args); err != nil {
			logger.Error(events.MessageEvent{Message: err.Error()})
			os.Exit(1)
		}
	}
}
