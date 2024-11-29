package cmd

import (
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/utils/cmdexec"
	"github.com/spf13/cobra"
)

var cmdConfig = &amigoconfig.Config{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:                "amigo",
	Short:              "Tool to manage database migrations with go files",
	SilenceUsage:       true,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		shellPath := amigoconfig.DefaultShellPath
		defaultAmigoFolder := amigoconfig.DefaultAmigoFolder

		if env, ok := os.LookupEnv("AMIGO_FOLDER"); ok {
			defaultAmigoFolder = env
		}

		schemaVersionTable := amigoconfig.DefaultSchemaVersionTable
		mainFilePath := path.Join(defaultAmigoFolder, "main.go")
		mainBinaryPath := path.Join(defaultAmigoFolder, "main")
		migrationFolder := amigoconfig.DefaultMigrationFolder

		if slices.Contains(args, "init") {
			return executeInit(mainFilePath, defaultAmigoFolder, schemaVersionTable, migrationFolder)
		}

		return executeMain(shellPath, mainFilePath, mainBinaryPath, args)
	},
}

func Execute() {
	_ = rootCmd.Execute()
}

func executeMain(shellPath, mainFilePath, mainBinaryPath string, restArgs []string) error {
	_, err := os.Stat(mainFilePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s file does not exist, please run 'amigo init' to create it", mainFilePath)
	}

	// build binary
	args := []string{
		"go", "build",
		"-o", mainBinaryPath,
		mainFilePath,
	}

	err = cmdexec.ExecToWriter(shellPath, []string{"-c", strings.Join(args, " ")}, nil, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	args = []string{
		"./" + mainBinaryPath,
	}

	if len(restArgs) > 0 {
		args = append(args, restArgs...)
	}

	err = cmdexec.ExecToWriter(shellPath, []string{"-c", strings.Join(args, " ")}, nil, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("%s throw an error: %w", mainFilePath, err)
	}

	return nil
}
