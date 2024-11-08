package cmd

import (
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoctx"
	"github.com/spf13/cobra"
)

// migrateCmd represents the up command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply the database",
	Run: wrapCobraFunc(func(cmd *cobra.Command, am amigo.Amigo, args []string) error {
		if err := cmdCtx.ValidateDSN(); err != nil {
			return err
		}

		if err := am.ExecuteMain(amigo.MainArgMigrate); err != nil {
			return err
		}

		return nil
	}),
}

// rollbackCmd represents the down command
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback the database",
	Run: wrapCobraFunc(func(cmd *cobra.Command, am amigo.Amigo, args []string) error {
		if err := cmdCtx.ValidateDSN(); err != nil {
			return err
		}

		return am.ExecuteMain(amigo.MainArgRollback)
	}),
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(migrateCmd)

	registerBase := func(cmd *cobra.Command, m *amigoctx.Migration) {
		cmd.Flags().StringVar(&m.Version, "version", "",
			"Apply a specific version format: 20240502083700 or 20240502083700_name.go")
		cmd.Flags().BoolVar(&m.DryRun, "dry-run", false, "Run the migrations without applying them")
		cmd.Flags().BoolVar(&m.ContinueOnError, "continue-on-error", false,
			"Will not rollback the migration if an error occurs")
		cmd.Flags().DurationVar(&m.Timeout, "timeout", amigoctx.DefaultTimeout, "The timeout for the migration")
		cmd.Flags().BoolVarP(&m.DumpSchemaAfter, "dump-schema-after", "d", false,
			"Dump schema after migrate/rollback (not compatible with --use-schema-dump)")
	}

	registerBase(migrateCmd, cmdCtx.Migration)
	migrateCmd.Flags().BoolVar(&cmdCtx.Migration.UseSchemaDump, "use-schema-dump", false,
		"Use the schema file to apply the migration (for fresh install without any migration)")

	registerBase(rollbackCmd, cmdCtx.Migration)
	rollbackCmd.Flags().IntVar(&cmdCtx.Migration.Steps, "steps", 1, "The number of steps to rollback")

}
