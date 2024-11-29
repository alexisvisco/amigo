package entrypoint

import (
	"context"
	"fmt"
	"os"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/spf13/cobra"
)

// migrateCmd represents the up command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply the database",
	Run: wrapCobraFunc(func(cmd *cobra.Command, am amigo.Amigo, args []string) error {
		if err := config.ValidateDSN(); err != nil {
			return err
		}

		db, err := database(*am.Config)
		if err != nil {
			return fmt.Errorf("unable to get database: %w", err)
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), am.Config.Migration.Timeout)
		defer cancelFunc()

		err = am.RunMigrations(amigo.RunMigrationParams{
			DB:         db,
			Direction:  types.MigrationDirectionUp,
			Migrations: migrations,
			LogOutput:  os.Stdout,
			Context:    ctx,
		})

		if err != nil {
			return fmt.Errorf("failed to migrate database: %w", err)
		}

		return nil
	}),
}

// rollbackCmd represents the down command
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback the database",
	Run: wrapCobraFunc(func(cmd *cobra.Command, am amigo.Amigo, args []string) error {
		if err := config.ValidateDSN(); err != nil {
			return err
		}

		db, err := database(*am.Config)
		if err != nil {
			return fmt.Errorf("unable to get database: %w", err)
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), am.Config.Migration.Timeout)
		defer cancelFunc()

		err = am.RunMigrations(amigo.RunMigrationParams{
			DB:         db,
			Direction:  types.MigrationDirectionDown,
			Migrations: migrations,
			LogOutput:  os.Stdout,
			Context:    ctx,
		})

		if err != nil {
			return fmt.Errorf("failed to migrate database: %w", err)
		}

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(migrateCmd)

	registerBase := func(cmd *cobra.Command, m *amigoconfig.MigrationConfig) {
		cmd.Flags().StringVar(&m.Version, "version", "",
			"Apply a specific version format: 20240502083700 or 20240502083700_name.go")
		cmd.Flags().BoolVar(&m.DryRun, "dry-run", false, "Run the migrations without applying them")
		cmd.Flags().BoolVar(&m.ContinueOnError, "continue-on-error", false,
			"Will not rollback the migration if an error occurs")
		cmd.Flags().DurationVar(&m.Timeout, "timeout", amigoconfig.DefaultTimeout, "The timeout for the migration")
		cmd.Flags().BoolVarP(&m.DumpSchemaAfter, "dump-schema-after", "d", false,
			"Dump schema after migrate/rollback (not compatible with --use-schema-dump)")
	}

	registerBase(migrateCmd, config.Migration)
	migrateCmd.Flags().BoolVar(&config.Migration.UseSchemaDump, "use-schema-dump", false,
		"Use the schema file to apply the migration (for fresh install without any migration)")

	registerBase(rollbackCmd, config.Migration)
	rollbackCmd.Flags().IntVar(&config.Migration.Steps, "steps", 1, "The number of steps to rollback")

}
