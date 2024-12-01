package entrypoint

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/colors"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Status explain the current state of the database.",
	Run: wrapCobraFunc(func(cmd *cobra.Command, am amigo.Amigo, args []string) error {
		if err := config.ValidateDSN(); err != nil {
			return err
		}

		db, migrations, err := provider(*am.Config)
		if err != nil {
			return fmt.Errorf("unable to get provided resources from main: %w", err)
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), am.Config.Migration.Timeout)
		defer cancelFunc()

		versions, err := am.GetStatus(ctx, db)
		if err != nil {
			return fmt.Errorf("unable to get status: %w", err)
		}

		hasVersion := func(version string) bool {
			for _, v := range versions {
				if v == version {
					return true
				}
			}
			return false
		}

		// show status of 10 last migrations
		b := &strings.Builder{}
		tw := tabwriter.NewWriter(b, 2, 0, 1, ' ', 0)

		defaultMigrations := sliceArrayOrDefault(migrations, 10)
		for i, m := range defaultMigrations {
			key := fmt.Sprintf("(%s) %s", m.Date().UTC().Format(utils.FormatTime), m.Name())
			value := colors.Red("not applied")
			if hasVersion(m.Date().UTC().Format(utils.FormatTime)) {
				value = colors.Green("applied")
			}
			fmt.Fprintf(tw, "%s\t\t%s", key, value)
			if i != len(defaultMigrations)-1 {
				fmt.Fprintln(tw)
			}
		}
		tw.Flush()
		logger.Info(events.MessageEvent{Message: b.String()})

		return nil
	}),
}

func sliceArrayOrDefault[T any](array []T, x int) []T {
	defaultMigrations := array
	if len(array) >= x {
		defaultMigrations = array[len(array)-x:]
	}
	return defaultMigrations
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
