package cmd

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/templates"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/spf13/cobra"
	"path"
	"time"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize migrations folder and add the first migration file",
	Run: wrapCobraFunc(func(cmd *cobra.Command, am amigo.Amigo, args []string) error {
		if err := cmdCtx.ValidateDSN(); err != nil {
			return err
		}

		// create the main file
		logger.Info(events.FolderAddedEvent{FolderName: cmdCtx.MigrationFolder})

		file, err := utils.CreateOrOpenFile(path.Join(cmdCtx.AmigoFolderPath, "main.go"))
		if err != nil {
			return fmt.Errorf("unable to open main.go file: %w", err)
		}

		err = am.GenerateMainFile(file)
		if err != nil {
			return err
		}

		logger.Info(events.FileAddedEvent{FileName: path.Join(cmdCtx.AmigoFolderPath, "main.go")})

		// create the base schema version table
		now := time.Now()
		migrationFileName := fmt.Sprintf("%s_create_table_schema_version.go", now.UTC().Format(utils.FormatTime))
		file, err = utils.CreateOrOpenFile(path.Join(cmdCtx.MigrationFolder, migrationFileName))
		if err != nil {
			return fmt.Errorf("unable to open migrations.go file: %w", err)
		}

		inUp, err := templates.GetInitCreateTableTemplate(templates.CreateTableData{Name: cmdCtx.SchemaVersionTable},
			am.Driver == types.DriverUnknown)
		if err != nil {
			return err
		}

		err = am.GenerateMigrationFile(&amigo.GenerateMigrationFileParams{
			Name:            "create_table_schema_version",
			Up:              inUp,
			Down:            "// nothing to do to keep the schema version table",
			Type:            types.MigrationFileTypeClassic,
			Now:             now,
			Writer:          file,
			UseSchemaImport: am.Driver != types.DriverUnknown,
			UseFmtImport:    am.Driver == types.DriverUnknown,
		})
		if err != nil {
			return err
		}
		logger.Info(events.FileAddedEvent{FileName: path.Join(cmdCtx.MigrationFolder, migrationFileName)})

		// create the migrations file where all the migrations will be stored
		file, err = utils.CreateOrOpenFile(path.Join(cmdCtx.MigrationFolder, migrationsFile))
		if err != nil {
			return err
		}

		err = am.GenerateMigrationsFiles(file)
		if err != nil {
			return err
		}

		logger.Info(events.FileAddedEvent{FileName: path.Join(cmdCtx.MigrationFolder, migrationsFile)})

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(initCmd)
}
