package entrypoint

import (
	"fmt"
	"path"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Dump the schema of the database using appropriate tool",
	Long: `Dump the schema of the database using appropriate tool.
Supported databases:
	- postgres with pg_dump`,
	Run: wrapCobraFunc(func(cmd *cobra.Command, am amigo.Amigo, args []string) error {
		if err := config.ValidateDSN(); err != nil {
			return err
		}

		return dumpSchema(am)
	}),
}

func dumpSchema(am amigo.Amigo) error {
	file, err := utils.CreateOrOpenFile(config.SchemaOutPath)
	if err != nil {
		return fmt.Errorf("unable to open/create file: %w", err)
	}

	defer file.Close()

	err = am.DumpSchema(file, false)
	if err != nil {
		return fmt.Errorf("unable to dump schema: %w", err)
	}

	logger.Info(events.FileModifiedEvent{FileName: path.Join(config.SchemaOutPath)})
	return nil
}

func init() {
	rootCmd.AddCommand(schemaCmd)
}
