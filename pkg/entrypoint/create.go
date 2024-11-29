package entrypoint

import (
	"bytes"
	"context"
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

		if err := config.ValidateDSN(); err != nil {
			return err
		}

		if err := config.Create.ValidateType(); err != nil {
			return err
		}

		inUp := ""

		if config.Create.Dump {
			buffer := &bytes.Buffer{}
			err := am.DumpSchema(buffer, true)
			if err != nil {
				return err
			}

			inUp += fmt.Sprintf("s.Exec(`%s`)\n", buffer.String())

			config.Create.Type = "classic"
		}

		now := time.Now()
		version := now.UTC().Format(utils.FormatTime)
		config.Create.Version = version

		ext := "go"
		if config.Create.Type == "sql" {
			ext = "sql"
		}
		migrationFileName := fmt.Sprintf("%s_%s.%s", version, flect.Underscore(args[0]), ext)
		file, err := utils.CreateOrOpenFile(filepath.Join(config.MigrationFolder, migrationFileName))
		if err != nil {
			return fmt.Errorf("unable to open/create  file: %w", err)
		}

		err = am.GenerateMigrationFile(&amigo.GenerateMigrationFileParams{
			Name:   args[0],
			Up:     inUp,
			Down:   "",
			Type:   types.MigrationFileType(config.Create.Type),
			Now:    now,
			Writer: file,
		})
		if err != nil {
			return err
		}

		logger.Info(events.FileAddedEvent{FileName: filepath.Join(config.MigrationFolder, migrationFileName)})

		// create the migrations file where all the migrations will be stored
		file, err = utils.CreateOrOpenFile(path.Join(config.MigrationFolder, migrationsFile))
		if err != nil {
			return err
		}

		err = am.GenerateMigrationsFiles(file)
		if err != nil {
			return err
		}

		logger.Info(events.FileModifiedEvent{FileName: path.Join(config.MigrationFolder, migrationsFile)})

		if config.Create.Skip {
			db, err := database(*am.Config)
			if err != nil {
				return fmt.Errorf("unable to get database: %w", err)
			}

			ctx, cancelFunc := context.WithTimeout(context.Background(), am.Config.Migration.Timeout)
			defer cancelFunc()
			err = am.SkipMigrationFile(ctx, db)
			if err != nil {
				return err
			}
		}

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVar(&config.Create.Type, "type", "change",
		"The type of migration to create, possible values are [classic, change, sql]")

	createCmd.Flags().BoolVarP(&config.Create.Dump, "dump", "d", false,
		"dump with pg_dump the current schema and add it to the current migration")

	createCmd.Flags().StringVar(&config.Create.SQLSeparator, "sql-separator", "-- migrate:down",
		"the separator to split the up and down part of the migration")

	createCmd.Flags().BoolVar(&config.Create.Skip, "skip", false,
		"skip will set the migration as applied without executing it")
}
