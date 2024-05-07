package cmd

import (
	"fmt"
	"github.com/alexisvisco/mig/pkg/mig"
	"github.com/alexisvisco/mig/pkg/templates"
	"github.com/alexisvisco/mig/pkg/utils/tracker"
	"github.com/spf13/cobra"
	"os"
	"path"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize migrations folder and add the first migration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("DSN", dsnFlag)
		t := tracker.NewLogger(jsonFlag, cmd.OutOrStdout())

		err := os.MkdirAll(migrationFolderFlag, 0755)
		if err != nil {
			return fmt.Errorf("unable to create migration folder: %w", err)
		}

		t.AddEvent(tracker.FolderAddedEvent{FolderName: migrationFolderFlag})

		err = os.MkdirAll(migFolderPathFlag, 0755)
		if err != nil {
			return fmt.Errorf("unable to create main folder: %w", err)
		}

		err = mig.GenerateMainFile(migFolderPathFlag, migrationFolderFlag)
		if err != nil {
			return err
		}

		template, err := templates.GetInitCreateTableTemplate(templates.CreateTableData{Name: schemaVersionTableFlag})
		if err != nil {
			return err
		}

		file, _, err := mig.GenerateMigrationFile(mig.GenerateMigrationFileOptions{
			Name:    "schema_version",
			Folder:  migrationFolderFlag,
			Driver:  getDriver(),
			Package: packageFlag,
			MigType: "change",
			InUp:    template,
			InDown:  "",
		})
		if err != nil {
			return err
		}

		t.AddEvent(tracker.FileAddedEvent{FileName: file})

		err = mig.GenerateMigrationsFile(migrationFolderFlag, packageFlag,
			path.Join(migrationFolderFlag, migrationsFile))
		if err != nil {
			return err
		}

		t.
			AddEvent(tracker.FileAddedEvent{FileName: path.Join(migFolderPathFlag, "main.go")}).
			AddEvent(tracker.FileAddedEvent{FileName: path.Join(migrationFolderFlag, migrationsFile)}).
			Measure()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
