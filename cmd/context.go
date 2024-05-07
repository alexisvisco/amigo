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
	Long: `A context is a file inside the .mig folder that contains the flags that you use in the command line.
	
Example: 
	mig context --dsn "postgres://user:password@host:port/dbname?sslmode=disable"

This command will create a file .mig/context.yaml with the content:
	dsn: "postgres://user:password@host:port/dbname?sslmode=disable"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := viper.WriteConfig()
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// contextCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// contextCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
