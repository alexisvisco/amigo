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
	rollbackVersionFlag         string
	rollbackStepsFlag           int
	rollbackDryRunFlag          bool
	rollbackContinueOnErrorFlag bool
	rollbackTimeoutFlag         time.Duration
)

// rollbackCmd represents the down command
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateDSN(); err != nil {
			return err
		}

		t := tracker.NewLogger(jsonFlag, cmd.OutOrStdout())

		var version *string
		if migrateVersionFlag != "" {
			version = &rollbackVersionFlag
		}

		switch getDriver() {
		case "postgres":
			return mig.ExecuteMain(path.Join(migFolderPathFlag, "main.go"), &mig.MainOptions{
				DSN:                dsnFlag,
				MigrationDirection: types.MigrationDirectionDown,
				Version:            version,
				SchemaVersionTable: schema.TableName(schemaVersionTableFlag),
				DryRun:             rollbackDryRunFlag,
				ContinueOnError:    rollbackContinueOnErrorFlag,
				Timeout:            rollbackTimeoutFlag,
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
	rootCmd.AddCommand(rollbackCmd)
	rollbackCmd.Flags().StringVar(&migrateVersionFlag, "version", "",
		"Apply a specific version format: 20240502083700 or 20240502083700_name.go")
	rollbackCmd.Flags().IntVar(&rollbackStepsFlag, "steps", 1, "The number of steps to rollback")
	rollbackCmd.Flags().BoolVar(&rollbackDryRunFlag, "dry-run", false, "Run the migrations without applying them")
	rollbackCmd.Flags().BoolVar(&rollbackContinueOnErrorFlag, "continue-on-error", false,
		"Will not rollback the migration if an error occurs")
	rollbackCmd.Flags().DurationVar(&rollbackTimeoutFlag, "timeout", 2*time.Minute, "The timeout for the migration")

}
