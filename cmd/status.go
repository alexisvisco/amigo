package cmd

import (
	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Status explain the current state of the database.",
	Run: wrapCobraFunc(func(cmd *cobra.Command, am amigo.Amigo, args []string) error {
		if err := cmdCtx.ValidateDSN(); err != nil {
			return err
		}

		return am.ExecuteMain(amigo.MainArgStatus)
	}),
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
