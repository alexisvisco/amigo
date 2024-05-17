package cmd

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoctx"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/spf13/viper"
	"os"
	"path/filepath"

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
	
	will create a folder named .amigo with a context file inside to not have to pass the dsn every time.
	$ amigo context --dsn "postgres://user:password@host:port/dbname?sslmode=disable"
	
	
	$ amigo init
	note: will create:
	- folder named migrations with a file named migrations.go that contains the list of migrations
	- a new migration to create the  schema version table
	- a main.go in the .amigo folder
	
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
	cmdCtx.Register(rootCmd)
	initConfig()
}

func initConfig() {
	// check if the file exists, if the file does not exist, create it
	if _, err := os.Stat(filepath.Join(cmdCtx.AmigoFolderPath, contextFileName)); os.IsNotExist(err) {
		err := os.MkdirAll(cmdCtx.AmigoFolderPath, 0755)
		if err != nil {
			fmt.Println("error: can't create folder:", err)
			os.Exit(1)
		}
		if err := viper.WriteConfigAs(filepath.Join(cmdCtx.AmigoFolderPath, contextFileName)); err != nil {
			fmt.Println("error: can't write config:", err)
			os.Exit(1)
		}
	}

	_ = viper.BindPFlag("dsn", rootCmd.PersistentFlags().Lookup("dsn"))
	_ = viper.BindPFlag("amigo-folder", rootCmd.PersistentFlags().Lookup("amigo-folder"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	_ = viper.BindPFlag("folder", rootCmd.PersistentFlags().Lookup("folder"))
	_ = viper.BindPFlag("package", rootCmd.PersistentFlags().Lookup("package"))
	_ = viper.BindPFlag("schema-version-table", rootCmd.PersistentFlags().Lookup("schema-version-table"))
	_ = viper.BindPFlag("shell-path", rootCmd.PersistentFlags().Lookup("shell-path"))
	_ = viper.BindPFlag("pg-dump-path", createCmd.Flags().Lookup("pg-dump-path"))
	_ = viper.BindPFlag("sql", rootCmd.PersistentFlags().Lookup("sql"))
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	viper.SetConfigFile(filepath.Join(cmdCtx.AmigoFolderPath, contextFileName))

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("error: can't read config:", err)
		os.Exit(1)
	}

	if viper.IsSet("dsn") {
		cmdCtx.DSN = viper.GetString("dsn")
	}

	if viper.IsSet("amigo-folder") {
		cmdCtx.AmigoFolderPath = viper.GetString("amigo-folder")
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

	if viper.IsSet("debug") {
		cmdCtx.Debug = viper.GetBool("debug")
	}
}

func getDriver(_ string) types.Driver {
	// TODO: implement this when we have more drivers
	return "postgres"
}

func wrapCobraFunc(f func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		amigo.SetupSlog(cmdCtx.ShowSQL, cmdCtx.Debug, cmdCtx.JSON, os.Stdout)

		if err := f(cmd, args); err != nil {
			logger.Error(events.MessageEvent{Message: err.Error()})
			// os.Exit(1)
		}
	}
}
