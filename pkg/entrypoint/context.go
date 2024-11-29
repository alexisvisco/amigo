package entrypoint

import (
	"fmt"
	"path/filepath"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const contextsFileName = "contexts.yml"

// contextCmd represents the context command
var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "show the current context yaml file",
	Long: `A context is a file inside the amigo folder that contains the flags that you use in the command line.
	
Example: 
	amigo context --dsn "postgres://user:password@host:port/dbname?sslmode=disable"

This command will create a file $amigo_folder/context.yaml with the content:
	dsn: "postgres://user:password@host:port/dbname?sslmode=disable"
`,
	Run: wrapCobraFunc(func(cmd *cobra.Command, a amigo.Amigo, args []string) error {
		content, err := utils.GetFileContent(filepath.Join(a.Config.AmigoFolderPath, contextsFileName))
		if err != nil {
			return fmt.Errorf("unable to read contexts file: %w", err)
		}

		fmt.Println(string(content))

		return nil
	}),
}

var ContextSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the current context",
	Run: wrapCobraFunc(func(cmd *cobra.Command, a amigo.Amigo, args []string) error {
		yamlConfig, err := amigoconfig.LoadYamlConfig(filepath.Join(a.Config.AmigoFolderPath, contextsFileName))
		if err != nil {
			return fmt.Errorf("unable to read contexts file: %w", err)
		}

		if len(args) == 0 {
			return fmt.Errorf("missing context name")
		}

		if _, ok := yamlConfig.Contexts[args[0]]; !ok {
			return fmt.Errorf("context %s not found", args[0])
		}

		yamlConfig.CurrentContext = args[0]

		file, err := utils.CreateOrOpenFile(filepath.Join(a.Config.AmigoFolderPath, contextsFileName))
		if err != nil {
			return fmt.Errorf("unable to open contexts file: %w", err)
		}
		defer file.Close()

		err = file.Truncate(0)
		if err != nil {
			return fmt.Errorf("unable to truncate contexts file: %w", err)
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			return fmt.Errorf("unable to seek file: %w", err)
		}

		yamlOut, err := yaml.Marshal(yamlConfig)
		if err != nil {
			return fmt.Errorf("unable to marshal yaml: %w", err)
		}

		_, err = file.WriteString(string(yamlOut))
		if err != nil {
			return fmt.Errorf("unable to write contexts file: %w", err)
		}

		logger.Info(events.FileModifiedEvent{FileName: filepath.Join(a.Config.AmigoFolderPath, contextsFileName)})
		logger.Info(events.MessageEvent{Message: "context set to " + args[0]})

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(ContextSetCmd)
}
