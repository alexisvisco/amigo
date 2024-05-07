package mig

import (
	"database/sql"
	"fmt"
	"github.com/alexisvisco/mig/pkg/schema"
	"github.com/alexisvisco/mig/pkg/types"
	"github.com/alexisvisco/mig/pkg/utils"
	"github.com/alexisvisco/mig/pkg/utils/cmdexec"
	"github.com/alexisvisco/mig/pkg/utils/tracker"
	"os"
	"path"
	"strings"
	"time"
)

type MainOptions struct {
	DSN        string
	Connection *sql.DB

	MigrationDirection types.MigrationDirection

	Version            *string
	Steps              *int
	SchemaVersionTable schema.TableName
	DryRun             bool
	ContinueOnError    bool
	Timeout            time.Duration
	Migrations         []schema.Migration

	JSON    bool
	Verbose bool

	Shell   string
	Tracker tracker.Tracker `json:"-"`
}

func ExecuteMain(mainLocation string, options *MainOptions) error {
	mainFolder := path.Dir(mainLocation)

	stat, err := os.Stat(mainFolder)
	if os.IsNotExist(err) {
		err := os.MkdirAll(mainFolder, 0755)
		if err != nil {
			return fmt.Errorf("unable to create main folder: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("unable to check if main folder exist: %w", err)
	}

	if stat != nil && !stat.IsDir() {
		return fmt.Errorf("main folder is not a directory")
	}

	mainFilePath := path.Join(mainFolder, "main.go")
	_, err = os.Stat(mainFilePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s file does not exist, please run 'mig init' to create it", mainFilePath)
	}

	args := []string{
		"go",
		"run",
		path.Join(mainFolder, "main.go"),
		"-dsn", options.DSN,
		"-direction", string(options.MigrationDirection),
		"-timeout", options.Timeout.String(),
		"-schema-version-table", string(options.SchemaVersionTable),
	}

	if options.Steps != nil {
		args = append(args, "-steps", fmt.Sprintf("%d", *options.Steps))
	}

	if options.Verbose {
		args = append(args, "-verbose")
	}

	if options.ContinueOnError {
		args = append(args, "-continue-on-error")
	}

	if options.DryRun {
		args = append(args, "-dry-run")
	}
	if options.Version != nil {
		v, err := utils.ParseMigrationVersion(*options.Version)
		if err != nil {
			return fmt.Errorf("unable to parse version: %w", err)
		}
		args = append(args, "-version", v)
	}

	if options.JSON {
		args = append(args, "-json")
	}

	if options.Verbose {
		options.Tracker.AddEvent(tracker.InfoEvent{Message: fmt.Sprintf("executing %s", args)})
	}

	err = cmdexec.ExecToWriter(options.Shell, []string{"-c", strings.Join(args, " ")}, nil, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("unable to execute %s: %w", mainFilePath, err)
	}

	return nil
}
