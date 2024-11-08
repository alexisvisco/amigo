package amigoctx

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/alexisvisco/amigo/pkg/types"
)

var (
	ErrDSNEmpty = errors.New("dsn is empty")
)

var (
	DefaultSchemaVersionTable = "public.mig_schema_versions"
	DefaultAmigoFolder        = "db"
	DefaultMigrationFolder    = "db/migrations"
	DefaultPackagePath        = "migrations"
	DefaultShellPath          = "/bin/bash"
	DefaultPGDumpPath         = "pg_dump"
	DefaultSchemaOutPath      = "db/schema.sql"
	DefaultTimeout            = 2 * time.Minute
	DefaultDBDumpSchema       = "public"
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
			SchemaOutPath:      DefaultSchemaOutPath,
			SchemaDBDumpSchema: DefaultDBDumpSchema,
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
	SchemaOutPath             string
	SchemaDBDumpSchema        string
	Debug                     bool
}

func (r *Context) GetRealDSN() string {
	switch types.GetDriver(r.Root.DSN) {
	case types.DriverSQLite:
		return strings.TrimPrefix(r.Root.DSN, "sqlite:")
	}

	return r.Root.DSN
}

func (a *Context) WithAmigoFolder(folder string) *Context {
	a.Root.AmigoFolderPath = folder
	return a
}

func (a *Context) WithMigrationFolder(folder string) *Context {
	a.Root.MigrationFolder = folder
	return a
}

func (a *Context) WithPackagePath(packagePath string) *Context {
	a.Root.PackagePath = packagePath
	return a
}

func (a *Context) WithSchemaVersionTable(table string) *Context {
	a.Root.SchemaVersionTable = table
	return a
}

func (a *Context) WithDSN(dsn string) *Context {
	a.Root.DSN = dsn
	return a
}

func (a *Context) WithVersion(version string) *Context {
	a.Migration.Version = version
	return a
}

func (a *Context) WithDumpSchemaAfterMigrating(dumpSchema bool) *Context {
	if a.Migration == nil {
		a.Migration = &Migration{}
	}
	a.Migration.DumpSchemaAfter = dumpSchema

	return a
}

func (a *Context) WithSteps(steps int) *Context {
	a.Migration.Steps = steps
	return a
}

func (a *Context) WithShowSQL(showSQL bool) *Context {
	a.Root.ShowSQL = showSQL
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
	UseSchemaDump   bool
	DumpSchemaAfter bool
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
	Type string
	Dump bool

	SQLSeparator string

	Skip bool
	// Version is post setted after the name have been generated from the arg and time
	Version string
}

func (c *Create) ValidateType() error {
	allowedTypes := []string{string(types.MigrationFileTypeClassic), string(types.MigrationFileTypeChange), string(types.MigrationFileTypeSQL)}

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

		if toMerge.Root.SchemaDBDumpSchema != "" {
			defaultCtx.Root.SchemaDBDumpSchema = toMerge.Root.SchemaDBDumpSchema
		}

		if toMerge.Root.SchemaOutPath != "" {
			defaultCtx.Root.SchemaOutPath = toMerge.Root.SchemaOutPath
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

		if toMerge.Migration.UseSchemaDump {
			defaultCtx.Migration.UseSchemaDump = toMerge.Migration.UseSchemaDump
		}
	}

	if toMerge.Create != nil {
		if toMerge.Create.Type != "" {
			defaultCtx.Create.Type = toMerge.Create.Type
		}

		if toMerge.Create.Dump {
			defaultCtx.Create.Dump = toMerge.Create.Dump
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
