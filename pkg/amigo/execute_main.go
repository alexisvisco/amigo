package amigo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/alexisvisco/amigo/pkg/utils/cmdexec"
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

	amigoCtxJson, err := json.Marshal(a.ctx)
	if err != nil {
		return fmt.Errorf("unable to marshal amigo context: %w", err)
	}

	bas64Json := base64.StdEncoding.EncodeToString(amigoCtxJson)
	args = []string{
		"./" + mainBinaryPath,
		"-json", bas64Json,
	}

	args = append(args, string(arg))

	err = cmdexec.ExecToWriter(a.ctx.ShellPath, []string{"-c", strings.Join(args, " ")}, nil, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("%s throw an error: %w", mainFilePath, err)
	}

	return nil
}
