package amigoctx

import (
	"errors"
	"fmt"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
	"time"
)

var (
	ErrDSNEmpty = errors.New("dsn is empty")
)

var (
	DefaultSchemaVersionTable = "public.mig_schema_versions"
	DefaultAmigoFolder        = ".amigo"
	DefaultMigrationFolder    = "migrations"
	DefaultPackagePath        = "migrations"
	DefaultShellPath          = "/bin/bash"
	DefaultPGDumpPath         = "pg_dump"
	DefaultTimeout            = 2 * time.Minute
)

type Context struct {
	*Root

	Migrate  *Migrate
	Rollback *Rollback
	Create   *Create
}

func NewContext() *Context {
	return &Context{
		Root:     &Root{},
		Migrate:  &Migrate{},
		Rollback: &Rollback{},
		Create:   &Create{},
	}
}

type Root struct {
	AmigoFolderPath    string
	DSN                string
	JSON               bool
	ShowSQL            bool
	MigrationFolder    string
	PackagePath        string
	SchemaVersionTable string
	ShellPath          string
	PGDumpPath         string
	Debug              bool
}

func (r *Root) Register(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&r.AmigoFolderPath, "amigo-folder", "m", DefaultAmigoFolder,
		"The folder path to use for creating amigo related files related to this repository")

	cmd.PersistentFlags().StringVar(&r.DSN, "dsn", "",
		"The database connection string example: postgres://user:password@host:port/dbname?sslmode=disable")

	cmd.PersistentFlags().BoolVarP(&r.JSON, "json", "j", false, "Output in json format")

	cmd.PersistentFlags().StringVar(&r.MigrationFolder, "folder", DefaultMigrationFolder,
		"The folder where the migrations are stored")

	cmd.PersistentFlags().StringVarP(&r.PackagePath, "package", "p", DefaultPackagePath,
		"The package name for the migrations")

	cmd.PersistentFlags().StringVarP(&r.SchemaVersionTable, "schema-version-table", "t",
		DefaultSchemaVersionTable, "The table name for the migrations")

	cmd.PersistentFlags().StringVar(&r.ShellPath, "shell-path", DefaultShellPath,
		"the shell to use (for: amigo create --dump, it uses pg dump command)")

	cmd.PersistentFlags().BoolVar(&r.ShowSQL, "sql", false, "Print SQL queries")

	cmd.Flags().StringVar(&r.PGDumpPath, "pg-dump-path", DefaultPGDumpPath,
		"the path to the pg_dump command if --dump is set")

	cmd.PersistentFlags().BoolVar(&r.Debug, "debug", false, "Print debug information")
}

func (r *Root) ValidateDSN() error {
	if r.DSN == "" {
		return ErrDSNEmpty
	}

	allowedDrivers := []string{"postgres"}

	for _, driver := range allowedDrivers {
		if strings.Contains(r.DSN, driver) {
			return nil
		}
	}

	return fmt.Errorf("unsupported driver, allowed drivers are: %s", strings.Join(allowedDrivers, ", "))
}

type Migrate struct {
	Version         string
	DryRun          bool
	ContinueOnError bool
	Timeout         time.Duration
}

func (m *Migrate) Register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&m.Version, "version", "",
		"Apply a specific version format: 20240502083700 or 20240502083700_name.go")
	cmd.Flags().BoolVar(&m.DryRun, "dry-run", false, "Run the migrations without applying them")
	cmd.Flags().BoolVar(&m.ContinueOnError, "continue-on-error", false,
		"Will not rollback the migration if an error occurs")
	cmd.Flags().DurationVar(&m.Timeout, "timeout", DefaultTimeout, "The timeout for the migration")
}

type Rollback struct {
	Version         string
	Steps           int
	DryRun          bool
	ContinueOnError bool
	Timeout         time.Duration
}

func (r *Rollback) Register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&r.Version, "version", "",
		"Apply a specific version format: 20240502083700 or 20240502083700_name.go")
	cmd.Flags().IntVar(&r.Steps, "steps", 1, "The number of steps to rollback")
	cmd.Flags().BoolVar(&r.DryRun, "dry-run", false, "Run the migrations without applying them")
	cmd.Flags().BoolVar(&r.ContinueOnError, "continue-on-error", false,
		"Will not rollback the migration if an error occurs")
	cmd.Flags().DurationVar(&r.Timeout, "timeout", DefaultTimeout, "The timeout for the migration")
}

func (r *Rollback) ValidateSteps() error {
	if r.Steps < 0 {
		return fmt.Errorf("steps must be greater than 0")
	}

	return nil
}

func (r *Rollback) ValidateVersion() error {
	if r.Version == "" {
		return nil
	}

	re := regexp.MustCompile(`\d{14}(_\w+)?\.go`)
	if !re.MatchString(r.Version) {
		return fmt.Errorf("version must be in the format: 20240502083700 or 20240502083700_name.go")
	}

	return nil
}

type Create struct {
	Type   string
	Dump   bool
	Schema string
	Skip   bool
}

func (c *Create) Register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.Type, "type", "change",
		"The type of migration to create, possible values are [classic, change]")

	cmd.Flags().BoolVarP(&c.Dump, "dump", "d", false,
		"dump with pg_dump the current schema and add it to the current migration")

	cmd.Flags().StringVarP(&c.Schema, "dump-schema", "s", "public", "the schema to dump if --dump is set")

	cmd.Flags().BoolVar(&c.Skip, "skip", false,
		"skip will set the migration as applied without executing it")
}

func (c *Create) ValidateType() error {
	allowedTypes := []string{string(types.MigrationFileTypeClassic), string(types.MigrationFileTypeChange)}

	for _, t := range allowedTypes {
		if c.Type == t {
			return nil
		}
	}

	return fmt.Errorf("unsupported type, allowed types are: %s", strings.Join(allowedTypes, ", "))
}
