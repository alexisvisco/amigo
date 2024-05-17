package amigo

import (
	"database/sql"
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/cmdexec"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"os"
	"path"
	"strings"
	"time"
)

type RunMigrationOptions struct {
	// DSN is the data source name. Example "postgres://user:password@localhost:5432/dbname?sslmode=disable"
	DSN string

	// Connection is the database connection to use. If Connection is set, DSN is ignored.
	Connection *sql.DB

	// MigrationDirection is the direction of the migration.
	// It can be either up or down.
	MigrationDirection types.MigrationDirection

	// Version is the version of the migration to apply.
	Version *string

	// Steps is the number of migrations to apply.
	Steps *int

	// SchemaVersionTable is the name of the table that will store the schema version.
	SchemaVersionTable schema.TableName

	// DryRun specifies if the migrator should perform the migrations without actually applying them.q
	DryRun bool

	// ContinueOnError specifies if the migrator should continue running migrations even if an error occurs.
	ContinueOnError bool

	// Timeout is the maximum time the migration can take.
	Timeout time.Duration

	// Migrations is the list of all existing migrations.
	Migrations []schema.Migration

	JSON    bool
	ShowSQL bool
	Debug   bool

	// Shell is the shell to use, only used by the CLI
	Shell string
}

func ExecuteMain(mainLocation string, options *RunMigrationOptions) error {
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
		return fmt.Errorf("%s file does not exist, please run 'amigo init' to create it", mainFilePath)
	}

	// build binary
	args := []string{
		"go", "build",
		"-o", path.Join(mainFolder, "main"),
		path.Join(mainFolder, "main.go"),
	}

	err = cmdexec.ExecToWriter(options.Shell, []string{"-c", strings.Join(args, " ")}, nil, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	args = []string{
		"./" + path.Join(mainFolder, "main"),
		"-dsn", options.DSN,
		"-direction", string(options.MigrationDirection),
		"-timeout", options.Timeout.String(),
		"-schema-version-table", string(options.SchemaVersionTable),
	}

	if options.Steps != nil {
		args = append(args, "-steps", fmt.Sprintf("%d", *options.Steps))
	}

	if options.ShowSQL {
		args = append(args, "-sql")
	}

	if options.ContinueOnError {
		args = append(args, "-continue-on-error")
	}

	if options.DryRun {
		args = append(args, "-dry-run")
	}

	if options.Debug {
		args = append(args, "-debug")
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

	if options.Debug {
		logger.Debug(events.MessageEvent{Message: fmt.Sprintf("executing %s", args)})
	}

	err = cmdexec.ExecToWriter(options.Shell, []string{"-c", strings.Join(args, " ")}, nil, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("%s throw an error: %w", mainFilePath, err)
	}

	return nil
}
