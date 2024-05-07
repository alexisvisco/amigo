package cmd

import (
	"fmt"
	"github.com/alexisvisco/mig/pkg/types"
	"github.com/spf13/viper"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	migFolderPathFlag      string
	dsnFlag                string
	jsonFlag               bool
	verboseFlag            bool
	migrationFolderFlag    string
	packageFlag            string
	schemaVersionTableFlag string
	shellPathFlag          string
	pgDumpPathFlag         string
)

const (
	migrationsFile = "migrations.go"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mig",
	Short: "Tool to manage database migrations with go files",
	Long: `Basic usage: 
First you need to create a main folder with mig init:
	
	will create a folder named .mig with a context file inside to not have to pass the dsn every time.
	$ mig context --dsn "postgres://user:password@host:port/dbname?sslmode=disable"
	
	
	$ mig init
	note: will create:
	- folder named migrations with a file named migrations.go that contains the list of migrations
	- a new migration to create the  schema version table
	- a main.go in the .mig folder
	
Apply migrations:
	$ mig migrate
	note: you can set --version <version> to migrate a specific version

Create a new migration:
	$ mig create "create_table_users"
	note: you can set --dump if you already have a database and you want to create the first migration with what's 
	already in the database. --skip will add the version of the created migration inside the schema version table.

Rollback a migration:
	$ mig rollback
	note: you can set --step <number> to rollback a specific number of migrations, and --version <version> to rollback 
	to a specific version
`,
	SilenceUsage: true,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&migFolderPathFlag, "mig-folder", "m", ".mig",
		"The folder path to use for creating mig related files related to this repository")

	rootCmd.PersistentFlags().StringVar(&dsnFlag, "dsn", "",
		"The database connection string example: postgres://user:password@host:port/dbname?sslmode=disable")

	rootCmd.PersistentFlags().BoolVarP(&jsonFlag, "json", "j", false, "Output in json format")

	rootCmd.PersistentFlags().StringVar(&migrationFolderFlag, "folder", "migrations",
		"The folder where the migrations are stored")

	rootCmd.PersistentFlags().StringVarP(&packageFlag, "package", "p", "migrations",
		"The package name for the migrations")

	rootCmd.PersistentFlags().StringVarP(&schemaVersionTableFlag, "schema-version-table", "t",
		"public.mig_schema_versions", "The table name for the migrations")

	rootCmd.PersistentFlags().StringVar(&shellPathFlag, "shell-path", "/bin/bash",
		"the shell to use (for: mig create --dump, it uses pg dump command)")

	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose output (print SQL queries)")

	createCmd.Flags().StringVar(&pgDumpPathFlag, "pg-dump-path", "pg_dump",
		"the path to the pg_dump command if --dump is set")

	initConfig()

}

func initConfig() {
	// check if the file exists, if the file does not exist, create it
	if _, err := os.Stat(filepath.Join(migFolderPathFlag, contextFileName)); os.IsNotExist(err) {
		if err := viper.WriteConfigAs(filepath.Join(migFolderPathFlag, contextFileName)); err != nil {
			fmt.Println("Can't write config:", err)
			os.Exit(1)
		}
	}

	_ = viper.BindPFlag("dsn", rootCmd.PersistentFlags().Lookup("dsn"))
	_ = viper.BindPFlag("mig-folder", rootCmd.PersistentFlags().Lookup("mig-folder"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	_ = viper.BindPFlag("folder", rootCmd.PersistentFlags().Lookup("folder"))
	_ = viper.BindPFlag("package", rootCmd.PersistentFlags().Lookup("package"))
	_ = viper.BindPFlag("schema-version-table", rootCmd.PersistentFlags().Lookup("schema-version-table"))
	_ = viper.BindPFlag("shell-path", rootCmd.PersistentFlags().Lookup("shell-path"))
	_ = viper.BindPFlag("pg-dump-path", createCmd.Flags().Lookup("pg-dump-path"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	viper.SetConfigFile(filepath.Join(migFolderPathFlag, contextFileName))

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}

	if viper.IsSet("dsn") {
		dsnFlag = viper.GetString("dsn")
	}

	if viper.IsSet("mig-folder") {
		migFolderPathFlag = viper.GetString("mig-folder")
	}

	if viper.IsSet("json") {
		jsonFlag = viper.GetBool("json")
	}

	if viper.IsSet("folder") {
		migrationFolderFlag = viper.GetString("folder")
	}

	if viper.IsSet("package") {
		packageFlag = viper.GetString("package")
	}

	if viper.IsSet("schema-version-table") {
		schemaVersionTableFlag = viper.GetString("schema-version-table")
	}

	if viper.IsSet("shell-path") {
		shellPathFlag = viper.GetString("shell-path")
	}

	if viper.IsSet("pg-dump-path") {
		pgDumpPathFlag = viper.GetString("pg-dump-path")
	}

	if viper.IsSet("verbose") {
		verboseFlag = viper.GetBool("verbose")
	}

}

func validateDSN() error {
	if dsnFlag == "" {
		return fmt.Errorf("dsn is required, example: postgres://user:password@host:port/dbname?sslmode=disable")
	}

	return nil
}

func getDriver() types.Driver {
	// TODO: implement this when we have more drivers
	return "postgres"
}
