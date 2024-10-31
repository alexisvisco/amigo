package cmd

import (
	"fmt"
	"path"
	"path/filepath"
	"time"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/gobuffalo/flect"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new migration in the migration folder",
	Run: wrapCobraFunc(func(cmd *cobra.Command, am amigo.Amigo, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("name is required: amigo create <name>")
		}

		if err := cmdCtx.ValidateDSN(); err != nil {
			return err
		}

		if err := cmdCtx.Create.ValidateType(); err != nil {
			return err
		}

		inUp := ""

		if cmdCtx.Create.Dump {
			dump, err := am.DumpSchema()
			if err != nil {
				return err
			}

			inUp += fmt.Sprintf("s.Exec(`%s`)\n", dump)

			cmdCtx.Create.Type = "classic"
		}

		now := time.Now()
		version := now.UTC().Format(utils.FormatTime)
		cmdCtx.Create.Version = version

		ext := "go"
		if cmdCtx.Create.Type == "sql" {
			ext = "sql"
		}
		migrationFileName := fmt.Sprintf("%s_%s.%s", version, flect.Underscore(args[0]), ext)
		file, err := utils.CreateOrOpenFile(filepath.Join(cmdCtx.MigrationFolder, migrationFileName))
		if err != nil {
			return fmt.Errorf("unable to open/create  file: %w", err)
		}

		err = am.GenerateMigrationFile(&amigo.GenerateMigrationFileParams{
			Name:   args[0],
			Up:     inUp,
			Down:   "",
			Type:   types.MigrationFileType(cmdCtx.Create.Type),
			Now:    now,
			Writer: file,
		})
		if err != nil {
			return err
		}

		logger.Info(events.FileAddedEvent{FileName: filepath.Join(cmdCtx.MigrationFolder, migrationFileName)})

		// create the migrations file where all the migrations will be stored
		file, err = utils.CreateOrOpenFile(path.Join(cmdCtx.MigrationFolder, migrationsFile))
		if err != nil {
			return err
		}

		err = am.GenerateMigrationsFiles(file)
		if err != nil {
			return err
		}

		logger.Info(events.FileModifiedEvent{FileName: path.Join(cmdCtx.MigrationFolder, migrationsFile)})

		if cmdCtx.Create.Skip {
			err = am.ExecuteMain(amigo.MainArgSkipMigration)
			if err != nil {
				return err
			}

			logger.Info(events.SkipMigrationEvent{MigrationVersion: version})
		}

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVar(&cmdCtx.Create.Type, "type", "change",
		"The type of migration to create, possible values are [classic, change, sql]")

	createCmd.Flags().BoolVarP(&cmdCtx.Create.Dump, "dump", "d", false,
		"dump with pg_dump the current schema and add it to the current migration")

	createCmd.Flags().StringVarP(&cmdCtx.Create.DumpSchema, "dump-schema", "s", "public",
		"the schema to dump if --dump is set")

	createCmd.Flags().StringVar(&cmdCtx.Create.SQLSeparator, "sql-separator", "-- migrate:down",
		"the separator to split the up and down part of the migration")

	createCmd.Flags().BoolVar(&cmdCtx.Create.Skip, "skip", false,
		"skip will set the migration as applied without executing it")
}
