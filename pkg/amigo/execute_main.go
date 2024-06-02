package amigo

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/cmdexec"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"os"
	"path"
	"strings"
)

type MainArg string

const (
	MainArgMigrate       MainArg = "migrate"
	MainArgRollback      MainArg = "rollback"
	MainArgSkipMigration MainArg = "skip-migration"
	MainArgStatus        MainArg = "status"
)

func (m MainArg) Validate() error {
	switch m {
	case MainArgMigrate, MainArgRollback, MainArgSkipMigration, MainArgStatus:
		return nil
	}

	return fmt.Errorf("invalid main arg: %s", m)
}

func (a Amigo) ExecuteMain(arg MainArg) error {
	mainFilePath := path.Join(a.ctx.AmigoFolderPath, "main.go")
	mainBinaryPath := path.Join(a.ctx.AmigoFolderPath, "main")
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

	err = cmdexec.ExecToWriter(a.ctx.ShellPath, []string{"-c", strings.Join(args, " ")}, nil, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	args = []string{
		"./" + mainBinaryPath,
		"-dsn", fmt.Sprintf(`"%s"`, a.ctx.DSN),
		"-schema-version-table", a.ctx.SchemaVersionTable,
	}

	if a.ctx.ShowSQL {
		args = append(args, "-sql")
	}

	if a.ctx.ShowSQLSyntaxHighlighting {
		args = append(args, "-sql-syntax-highlighting")
	}

	if a.ctx.Debug {
		args = append(args, "-debug")
	}

	if a.ctx.JSON {
		args = append(args, "-json")
	}

	if a.ctx.Debug {
		logger.Debug(events.MessageEvent{Message: fmt.Sprintf("executing %s", args)})
	}

	switch arg {
	case MainArgMigrate, MainArgRollback:
		if a.ctx.Migration.ContinueOnError {
			args = append(args, "-continue-on-error")
		}

		if a.ctx.Migration.DryRun {
			args = append(args, "-dry-run")
		}

		if a.ctx.Migration.Version != "" {
			v, err := utils.ParseMigrationVersion(a.ctx.Migration.Version)
			if err != nil {
				return fmt.Errorf("unable to parse version: %w", err)
			}
			args = append(args, "-version", v)
		}

		args = append(args, "-steps", fmt.Sprintf("%d", a.ctx.Migration.Steps))
	case MainArgSkipMigration:
		args = append(args, "-version", a.ctx.Create.Version)
	}

	args = append(args, string(arg))

	err = cmdexec.ExecToWriter(a.ctx.ShellPath, []string{"-c", strings.Join(args, " ")}, nil, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("%s throw an error: %w", mainFilePath, err)
	}

	return nil
}
