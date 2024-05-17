package cmd

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/spf13/cobra"
	"path/filepath"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new migration in the migration folder",
	Run: wrapCobraFunc(func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("name is required: amigo create <name>")
		}

		if err := cmdCtx.ValidateDSN(); err != nil {
			return err
		}

		if err := cmdCtx.Create.ValidateType(); err != nil {
			return err
		}

		migFileType := types.MigrationFileType(cmdCtx.Create.Type)

		inUp := ""

		if cmdCtx.Create.Dump {
			migFileType = types.MigrationFileTypeClassic

			dump, err := amigo.DumpSchema(&amigo.DumpSchemaOptions{
				DSN:                cmdCtx.DSN,
				MigrationTableName: cmdCtx.SchemaVersionTable,
				PGDumpPath:         cmdCtx.PGDumpPath,
				Schema:             cmdCtx.Create.Schema,
				Shell:              cmdCtx.ShellPath,
			})
			if err != nil {
				return err
			}

			inUp += fmt.Sprintf("s.Exec(`%s`)\n", dump)
		}

		filePath, version, err := amigo.GenerateMigrationFile(amigo.GenerateMigrationFileOptions{
			Name:    args[0],
			Folder:  cmdCtx.MigrationFolder,
			Driver:  getDriver(cmdCtx.DSN),
			Package: cmdCtx.PackagePath,
			MigType: migFileType,
			InUp:    inUp,
		})
		if err != nil {
			return err
		}

		logger.Info(events.FileAddedEvent{FileName: filePath})

		if cmdCtx.Create.Skip {
			connection, _, err := amigo.GetConnection(cmdCtx.DSN)
			if err != nil {
				return err
			}

			_, err = connection.Exec("INSERT INTO "+cmdCtx.SchemaVersionTable+" (version) VALUES ($1)", version)
			if err != nil {
				return fmt.Errorf("unable to set migration as applied: %w", err)
			}

			logger.Info(events.SkipMigrationEvent{MigrationVersion: version})
		}

		err = amigo.GenerateMigrationsFile(cmdCtx.MigrationFolder, cmdCtx.PackagePath,
			filepath.Join(cmdCtx.MigrationFolder, migrationsFile))
		if err != nil {
			return err
		}

		logger.Info(events.FileModifiedEvent{FileName: filepath.Join(cmdCtx.MigrationFolder, migrationsFile)})

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(createCmd)
	cmdCtx.Create.Register(createCmd)
}
