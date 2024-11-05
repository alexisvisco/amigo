package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoctx"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var cmdCtx = amigoctx.NewContext()

const (
	migrationsFile = "migrations.go"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "amigo",
	Short: "Tool to manage database migrations with go files",
	Long: `Basic usage: 
First you need to create a main folder with amigo init:
	
	will create a folder in migrations/db with a context file inside to not have to pass the dsn every time.
	
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

func Execute() {
	_ = rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cmdCtx.AmigoFolderPath, "amigo-folder", "m", amigoctx.DefaultAmigoFolder,
		"Folder path to use for creating amigo related files related to this repository")

	rootCmd.PersistentFlags().StringVar(&cmdCtx.DSN, "dsn", "",
		"The database connection string example: postgres://user:password@host:port/dbname?sslmode=disable")

	rootCmd.PersistentFlags().BoolVarP(&cmdCtx.JSON, "json", "j", false, "Output in json format")

	rootCmd.PersistentFlags().StringVar(&cmdCtx.MigrationFolder, "folder", amigoctx.DefaultMigrationFolder,
		"Folder where the migrations are stored")

	rootCmd.PersistentFlags().StringVarP(&cmdCtx.PackagePath, "package", "p", amigoctx.DefaultPackagePath,
		"Package name of the migrations folder")

	rootCmd.PersistentFlags().StringVarP(&cmdCtx.SchemaVersionTable, "schema-version-table", "t",
		amigoctx.DefaultSchemaVersionTable, "Table name to keep track of the migrations")

	rootCmd.PersistentFlags().StringVar(&cmdCtx.ShellPath, "shell-path", amigoctx.DefaultShellPath,
		"Shell to use (for: amigo create --dump, it uses pg dump command)")

	rootCmd.PersistentFlags().BoolVar(&cmdCtx.ShowSQL, "sql", false, "Print SQL queries")

	rootCmd.PersistentFlags().BoolVar(&cmdCtx.ShowSQLSyntaxHighlighting, "sql-syntax-highlighting", true,
		"Print SQL queries with syntax highlighting")

	rootCmd.Flags().StringVar(&cmdCtx.PGDumpPath, "pg-dump-path", amigoctx.DefaultPGDumpPath,
		"Path to the pg_dump command if --dump is set")

	rootCmd.PersistentFlags().BoolVar(&cmdCtx.Debug, "debug", false, "Print debug information")
	initConfig()
}

func initConfig() {
	// check if the file exists, if the file does not exist, create it
	if _, err := os.Stat(filepath.Join(cmdCtx.AmigoFolderPath, contextFileName)); os.IsNotExist(err) {
		err := os.MkdirAll(cmdCtx.AmigoFolderPath, 0755)
		if err != nil {
			logger.Error(events.MessageEvent{Message: fmt.Sprintf("error: can't create folder: %s", err)})
			os.Exit(1)
		}
		if err := viper.WriteConfigAs(filepath.Join(cmdCtx.AmigoFolderPath, contextFileName)); err != nil {
			logger.Error(events.MessageEvent{Message: fmt.Sprintf("error: can't write config: %s", err)})
			os.Exit(1)
		}
	}

	_ = viper.BindPFlag("dsn", rootCmd.PersistentFlags().Lookup("dsn"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	_ = viper.BindPFlag("folder", rootCmd.PersistentFlags().Lookup("folder"))
	_ = viper.BindPFlag("package", rootCmd.PersistentFlags().Lookup("package"))
	_ = viper.BindPFlag("schema-version-table", rootCmd.PersistentFlags().Lookup("schema-version-table"))
	_ = viper.BindPFlag("shell-path", rootCmd.PersistentFlags().Lookup("shell-path"))
	_ = viper.BindPFlag("pg-dump-path", createCmd.Flags().Lookup("pg-dump-path"))
	_ = viper.BindPFlag("sql", rootCmd.PersistentFlags().Lookup("sql"))
	_ = viper.BindPFlag("sql-syntax-highlighting", rootCmd.PersistentFlags().Lookup("sql-syntax-highlighting"))

	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	viper.SetConfigFile(filepath.Join(cmdCtx.AmigoFolderPath, contextFileName))

	if err := viper.ReadInConfig(); err != nil {
		logger.Error(events.MessageEvent{Message: fmt.Sprintf("error: can't read config: %s", err)})
		os.Exit(1)
	}

	if viper.IsSet("dsn") {
		cmdCtx.DSN = viper.GetString("dsn")
	}

	if viper.IsSet("json") {
		cmdCtx.JSON = viper.GetBool("json")
	}

	if viper.IsSet("folder") {
		cmdCtx.MigrationFolder = viper.GetString("folder")
	}

	if viper.IsSet("package") {
		cmdCtx.PackagePath = viper.GetString("package")
	}

	if viper.IsSet("schema-version-table") {
		cmdCtx.SchemaVersionTable = viper.GetString("schema-version-table")
	}

	if viper.IsSet("shell-path") {
		cmdCtx.ShellPath = viper.GetString("shell-path")
	}

	if viper.IsSet("pg-dump-path") {
		cmdCtx.PGDumpPath = viper.GetString("pg-dump-path")
	}

	if viper.IsSet("sql") {
		cmdCtx.ShowSQL = viper.GetBool("sql")
	}

	if viper.IsSet("sql-syntax-highlighting") {
		cmdCtx.ShowSQLSyntaxHighlighting = viper.GetBool("sql-syntax-highlighting")
	}

	if viper.IsSet("debug") {
		cmdCtx.Debug = viper.GetBool("debug")
	}
}

func wrapCobraFunc(f func(cmd *cobra.Command, am amigo.Amigo, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		am := amigo.NewAmigo(cmdCtx)
		am.SetupSlog(os.Stdout)

		if err := f(cmd, am, args); err != nil {
			logger.Error(events.MessageEvent{Message: err.Error()})
			os.Exit(1)
		}
	}
}
