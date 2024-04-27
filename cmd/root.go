package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// rootDSNFlag is the flag for the database connection string
	rootDSNFlag string
	jsonFlag    bool
	folderFlag  string
	packageFlag string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "mig",
	Short:        "mig is a tool to manage database migrations",
	SilenceUsage: true,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().StringVar(&rootDSNFlag, "dsn", "",
		"The database connection string example: postgres://user:password@host:port/dbname?sslmode=disable")

	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output in json format")

	rootCmd.PersistentFlags().StringVar(&folderFlag, "folder", "migrations",
		"the folder where the migrations are stored")
	rootCmd.PersistentFlags().StringVar(&packageFlag, "package", "migrations", "the package name for the migrations")
}

func validateDSN() error {
	if rootDSNFlag == "" {
		return fmt.Errorf("dsn is required, example: postgres://user:password@host:port/dbname?sslmode=disable")
	}

	return nil
}
