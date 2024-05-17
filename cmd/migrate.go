package cmd

import (
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/spf13/cobra"
	"path"
)

// migrateCmd represents the up command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply the database",
	Run: wrapCobraFunc(func(cmd *cobra.Command, args []string) error {
		if err := cmdCtx.ValidateDSN(); err != nil {
			return err
		}

		var version *string
		if cmdCtx.Migrate.Version != "" {
			version = &cmdCtx.Migrate.Version
		}

		return amigo.ExecuteMain(path.Join(cmdCtx.AmigoFolderPath, "main.go"), &amigo.RunMigrationOptions{
			DSN:                cmdCtx.DSN,
			MigrationDirection: types.MigrationDirectionUp,
			Version:            version,
			SchemaVersionTable: schema.TableName(cmdCtx.SchemaVersionTable),
			DryRun:             cmdCtx.Migrate.DryRun,
			ContinueOnError:    cmdCtx.Migrate.ContinueOnError,
			Timeout:            cmdCtx.Migrate.Timeout,
			JSON:               cmdCtx.JSON,
			Shell:              cmdCtx.ShellPath,
			ShowSQL:            cmdCtx.ShowSQL,
		})
	}),
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	cmdCtx.Migrate.Register(migrateCmd)
}
