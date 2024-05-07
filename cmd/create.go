package cmd

import (
	"fmt"
	"github.com/alexisvisco/mig/pkg/mig"
	"github.com/alexisvisco/mig/pkg/types"
	"github.com/alexisvisco/mig/pkg/utils"
	"github.com/alexisvisco/mig/pkg/utils/tracker"
	"github.com/spf13/cobra"
	"path/filepath"
)

var (
	createMigrationTypeFlag string
	createDumpFlag          bool
	createDumpSchema        string
	createSkipDump          bool
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

		migFileType := types.MigrationFileType(createMigrationTypeFlag)

		if !migFileType.IsValid() {
			return fmt.Errorf("invalid migration file type: %s, can be: %s", createMigrationTypeFlag,
				utils.StringJoin(types.MigrationFileTypeValues, ", "))
		}

		printer := tracker.NewLogger(jsonFlag, cmd.OutOrStdout())

		inUp := ""

		if createDumpFlag {
			migFileType = types.MigrationFileTypeClassic
			dump, err := mig.DumpSchema(&mig.DumpSchemaOptions{
				DSN:                dsnFlag,
				MigrationTableName: schemaVersionTableFlag,
				PGDumpPath:         pgDumpPathFlag,
				Schema:             createDumpSchema,
				Shell:              shellPathFlag,
			})
			if err != nil {
				return err
			}

			inUp += fmt.Sprintf("s.Exec(`%s`)\n", dump)
		}

		filePath, version, err := mig.GenerateMigrationFile(mig.GenerateMigrationFileOptions{
			Name:    args[0],
			Folder:  migrationFolderFlag,
			Driver:  getDriver(),
			Package: packageFlag,
			MigType: migFileType,
			InUp:    inUp,
		})
		if err != nil {
			return err
		}

		printer.AddEvent(tracker.FileAddedEvent{FileName: filePath})

		if createSkipDump {
			connection, err := mig.GetConnection(dsnFlag, verboseFlag)
			if err != nil {
				return err
			}

			_, err = connection.Exec("INSERT INTO "+schemaVersionTableFlag+" (version) VALUES ($1)", version)
			if err != nil {
				return fmt.Errorf("unable to set migration as applied: %w", err)
			}

			printer.AddEvent(tracker.SkipMigrationEvent{MigrationVersion: version})
		}

		err = mig.GenerateMigrationsFile(migrationFolderFlag, packageFlag,
			filepath.Join(migrationFolderFlag, migrationsFile))
		if err != nil {
			return err
		}

		printer.AddEvent(tracker.FileModifiedEvent{FileName: filepath.Join(migrationFolderFlag, migrationsFile)})

		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVar(&createMigrationTypeFlag, "type", "change",
		"The type of migration to create, possible values are [classic, change]")

	createCmd.Flags().BoolVarP(&createDumpFlag, "dump", "d", false,
		"dump with pg_dump the current schema and add it to the current migration")

	createCmd.Flags().StringVarP(&createDumpSchema, "dump-schema", "s", "public", "the schema to dump if --dump is set")

	createCmd.Flags().BoolVar(&createSkipDump, "skip", false,
		"skip will set the migration as applied without executing it")
}
