package cmd

import (
	"errors"
	"github.com/alexisvisco/mig/pkg/mig"
	"github.com/alexisvisco/mig/pkg/schema"
	"github.com/alexisvisco/mig/pkg/types"
	"github.com/alexisvisco/mig/pkg/utils/tracker"
	"github.com/spf13/cobra"
	"path"
	"time"
)

var (
	migrateVersionFlag         string
	migrateDryRunFlag          bool
	migrateContinueOnErrorFlag bool
	migrateTimeoutFlag         time.Duration
)

// migrateCmd represents the up command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateDSN(); err != nil {
			return err
		}

		t := tracker.NewLogger(jsonFlag, cmd.OutOrStdout())

		var version *string
		if migrateVersionFlag != "" {
			version = &migrateVersionFlag
		}

		switch getDriver() {
		case "postgres":
			return mig.ExecuteMain(path.Join(migFolderPathFlag, "main.go"), &mig.MainOptions{
				DSN:                dsnFlag,
				MigrationDirection: types.MigrationDirectionUp,
				Version:            version,
				SchemaVersionTable: schema.TableName(schemaVersionTableFlag),
				DryRun:             migrateDryRunFlag,
				ContinueOnError:    migrateContinueOnErrorFlag,
				Timeout:            migrateTimeoutFlag,
				JSON:               jsonFlag,
				Shell:              shellPathFlag,
				Verbose:            verboseFlag,
				Tracker:            t,
			})
		default:
			return errors.New("unsupported database")
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().StringVar(&migrateVersionFlag, "version", "",
		"Apply a specific version format: 20240502083700 or 20240502083700_name.go")
	migrateCmd.Flags().BoolVar(&migrateDryRunFlag, "dry-run", false, "Run the migrations without applying them")
	migrateCmd.Flags().BoolVar(&migrateContinueOnErrorFlag, "continue-on-error", false,
		"Will not rollback the migration if an error occurs")
	migrateCmd.Flags().DurationVar(&migrateTimeoutFlag, "timeout", 2*time.Minute, "The timeout for the migration")
}
