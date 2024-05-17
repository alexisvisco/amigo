package cmd

import (
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/spf13/cobra"
	"path"
)

// rollbackCmd represents the down command
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback the database",
	Run: wrapCobraFunc(func(cmd *cobra.Command, args []string) error {
		if err := cmdCtx.ValidateDSN(); err != nil {
			return err
		}

		var version *string
		if cmdCtx.Rollback.Version != "" {
			version = &cmdCtx.Rollback.Version
		}

		return amigo.ExecuteMain(path.Join(cmdCtx.AmigoFolderPath, "main.go"), &amigo.RunMigrationOptions{
			DSN:                cmdCtx.DSN,
			MigrationDirection: types.MigrationDirectionDown,
			Version:            version,
			SchemaVersionTable: schema.TableName(cmdCtx.SchemaVersionTable),
			DryRun:             cmdCtx.Rollback.DryRun,
			ContinueOnError:    cmdCtx.Rollback.ContinueOnError,
			Timeout:            cmdCtx.Rollback.Timeout,
			JSON:               cmdCtx.JSON,
			Shell:              cmdCtx.ShellPath,
			ShowSQL:            cmdCtx.ShowSQL,
		})
	}),
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
	cmdCtx.Rollback.Register(rollbackCmd)
}
