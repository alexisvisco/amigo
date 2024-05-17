package cmd

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/templates"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/spf13/cobra"
	"os"
	"path"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize migrations folder and add the first migration file",
	Run: wrapCobraFunc(func(cmd *cobra.Command, args []string) error {
		if err := cmdCtx.ValidateDSN(); err != nil {
			return err
		}

		err := os.MkdirAll(cmdCtx.MigrationFolder, 0755)
		if err != nil {
			return fmt.Errorf("unable to create migration folder: %w", err)
		}

		logger.Info(events.FolderAddedEvent{FolderName: cmdCtx.MigrationFolder})

		err = os.MkdirAll(cmdCtx.AmigoFolderPath, 0755)
		if err != nil {
			return fmt.Errorf("unable to create main folder: %w", err)
		}

		err = amigo.GenerateMainFile(cmdCtx.AmigoFolderPath, cmdCtx.MigrationFolder)
		if err != nil {
			return err
		}

		template, err := templates.GetInitCreateTableTemplate(templates.CreateTableData{Name: cmdCtx.SchemaVersionTable})
		if err != nil {
			return err
		}

		file, _, err := amigo.GenerateMigrationFile(amigo.GenerateMigrationFileOptions{
			Name:    "schema_version",
			Folder:  cmdCtx.MigrationFolder,
			Driver:  getDriver(cmdCtx.DSN),
			Package: cmdCtx.PackagePath,
			MigType: "change",
			InUp:    template,
			InDown:  "",
		})
		if err != nil {
			return err
		}

		logger.Info(events.FileAddedEvent{FileName: file})

		err = amigo.GenerateMigrationsFile(cmdCtx.MigrationFolder, cmdCtx.PackagePath,
			path.Join(cmdCtx.MigrationFolder, migrationsFile))
		if err != nil {
			return err
		}

		logger.Info(events.FileAddedEvent{FileName: path.Join(cmdCtx.AmigoFolderPath, "main.go")})
		logger.Info(events.FileAddedEvent{FileName: path.Join(cmdCtx.MigrationFolder, migrationsFile)})

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(initCmd)
}
