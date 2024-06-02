package amigoctx

import (
	"errors"
	"fmt"
	"github.com/alexisvisco/amigo/pkg/types"
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

	Migration *Migration
	Create    *Create
}

func NewContext() *Context {
	return &Context{
		Root: &Root{
			SchemaVersionTable: DefaultSchemaVersionTable,
			AmigoFolderPath:    DefaultAmigoFolder,
			MigrationFolder:    DefaultMigrationFolder,
			PackagePath:        DefaultPackagePath,
			ShellPath:          DefaultShellPath,
			PGDumpPath:         DefaultPGDumpPath,
		},
		Migration: &Migration{
			Timeout: DefaultTimeout,
			Steps:   1,
		},
		Create: &Create{},
	}
}

type Root struct {
	AmigoFolderPath           string
	DSN                       string
	JSON                      bool
	ShowSQL                   bool
	ShowSQLSyntaxHighlighting bool
	MigrationFolder           string
	PackagePath               string
	SchemaVersionTable        string
	ShellPath                 string
	PGDumpPath                string
	Debug                     bool
}

func (r *Context) GetRealDSN() string {
	switch types.GetDriver(r.Root.DSN) {
	case types.DriverSQLite:
		return strings.TrimPrefix(r.Root.DSN, "sqlite:")
	}

	return r.Root.DSN
}

func (a *Context) WithDSN(dsn string) *Context {
	a.Root.DSN = dsn
	return a
}

func (a *Context) WithVersion(version string) *Context {
	a.Migration.Version = version
	return a
}

func (a *Context) WithSteps(steps int) *Context {
	a.Migration.Steps = steps
	return a
}

func (r *Root) ValidateDSN() error {
	if r.DSN == "" {
		return ErrDSNEmpty
	}

	return nil
}

type Migration struct {
	Version         string
	Steps           int
	DryRun          bool
	ContinueOnError bool
	Timeout         time.Duration
}

func (m *Migration) ValidateVersion() error {
	if m.Version == "" {
		return nil
	}

	re := regexp.MustCompile(`\d{14}(_\w+)?\.go`)
	if !re.MatchString(m.Version) {
		return fmt.Errorf("version must be in the format: 20240502083700 or 20240502083700_name.go")
	}

	return nil
}

type Create struct {
	Type       string
	Dump       bool
	DumpSchema string
	Skip       bool

	// Version is post setted after the name have been generated from the arg and time
	Version string
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

func MergeContext(toMerge Context) *Context {
	defaultCtx := NewContext()

	if toMerge.Root != nil {
		if toMerge.Root.AmigoFolderPath != "" {
			defaultCtx.Root.AmigoFolderPath = toMerge.Root.AmigoFolderPath
		}

		if toMerge.Root.DSN != "" {
			defaultCtx.Root.DSN = toMerge.Root.DSN
		}

		if toMerge.Root.JSON {
			defaultCtx.Root.JSON = toMerge.Root.JSON
		}

		if toMerge.Root.ShowSQL {
			defaultCtx.Root.ShowSQL = toMerge.Root.ShowSQL
		}

		if toMerge.Root.MigrationFolder != "" {
			defaultCtx.Root.MigrationFolder = toMerge.Root.MigrationFolder
		}

		if toMerge.Root.PackagePath != "" {
			defaultCtx.Root.PackagePath = toMerge.Root.PackagePath
		}

		if toMerge.Root.SchemaVersionTable != "" {
			defaultCtx.Root.SchemaVersionTable = toMerge.Root.SchemaVersionTable
		}

		if toMerge.Root.ShellPath != "" {
			defaultCtx.Root.ShellPath = toMerge.Root.ShellPath
		}

		if toMerge.Root.PGDumpPath != "" {
			defaultCtx.Root.PGDumpPath = toMerge.Root.PGDumpPath
		}

		if toMerge.Root.Debug {
			defaultCtx.Root.Debug = toMerge.Root.Debug
		}
	}

	if toMerge.Migration != nil {
		if toMerge.Migration.Version != "" {
			defaultCtx.Migration.Version = toMerge.Migration.Version
		}

		if toMerge.Migration.Steps != 0 {
			defaultCtx.Migration.Steps = toMerge.Migration.Steps
		}

		if toMerge.Migration.DryRun {
			defaultCtx.Migration.DryRun = toMerge.Migration.DryRun
		}

		if toMerge.Migration.ContinueOnError {
			defaultCtx.Migration.ContinueOnError = toMerge.Migration.ContinueOnError
		}

		if toMerge.Migration.Timeout != 0 {
			defaultCtx.Migration.Timeout = toMerge.Migration.Timeout
		}
	}

	if toMerge.Create != nil {
		if toMerge.Create.Type != "" {
			defaultCtx.Create.Type = toMerge.Create.Type
		}

		if toMerge.Create.Dump {
			defaultCtx.Create.Dump = toMerge.Create.Dump
		}

		if toMerge.Create.DumpSchema != "" {
			defaultCtx.Create.DumpSchema = toMerge.Create.DumpSchema
		}

		if toMerge.Create.Skip {
			defaultCtx.Create.Skip = toMerge.Create.Skip
		}

		if toMerge.Create.Version != "" {
			defaultCtx.Create.Version = toMerge.Create.Version
		}
	}

	return defaultCtx
}
