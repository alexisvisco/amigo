package cmd

import (
	"fmt"
	"github.com/alexisvisco/mig/internal/cli"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"
)

var (
	createMigrationTypeFlag string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new migration in the migration folder",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("name is required: mig create <name>")
		}

		if err := validateDSN(); err != nil {
			return err
		}

		if !slices.Contains([]string{"classic", "change"}, createMigrationTypeFlag) {
			return fmt.Errorf("invalid migration type, possible values are [classic, change]")
		}

		printer := cli.NewPrinter()

		filePath, _, err := cli.CreateMigrationFile(cli.CreateMigrationFileOptions{
			Name:    args[0],
			Folder:  folderFlag,
			Driver:  "postgres",
			Package: packageFlag,
			MigType: createMigrationTypeFlag,
		})
		if err != nil {
			return err
		}

		printer.AddEvent(cli.FileAddedEvent{FileName: filePath})

		err = cli.GenerateMigrationsFile(folderFlag, packageFlag, filepath.Join(folderFlag, "migrations.go"))
		if err != nil {
			return err
		}

		printer.AddEvent(cli.FileModifiedEvent{FileName: filepath.Join(folderFlag, "migrations.go")})

		printer.Print(jsonFlag)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&createMigrationTypeFlag, "type", "t", "change",
		"The type of migration to create, possible values are [classic, change]")
}
