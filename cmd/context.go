package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const contextFileName = "config.yml"

// contextCmd represents the context command
var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "save flags into a context",
	Long: `A context is a file inside the .amigo folder that contains the flags that you use in the command line.
	
Example: 
	amigo context --dsn "postgres://user:password@host:port/dbname?sslmode=disable"

This command will create a file .amigo/context.yaml with the content:
	dsn: "postgres://user:password@host:port/dbname?sslmode=disable"
`,
	Run: wrapCobraFunc(func(cmd *cobra.Command, args []string) error {
		err := viper.WriteConfig()
		if err != nil {
			return err
		}

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(contextCmd)
}
