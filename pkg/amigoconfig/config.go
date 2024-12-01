package amigoconfig

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
	DefaultSchemaVersionTable          = "public.mig_schema_versions"
	DefaultAmigoFolder                 = "db"
	DefaultMigrationFolder             = "db/migrations"
	DefaultMigrationPackageName        = "migrations"
	DefaultShellPath                   = "/bin/bash"
	DefaultPGDumpPath                  = "pg_dump"
	DefaultSchemaOutPath               = "db/schema.sql"
	DefaultTimeout                     = 2 * time.Minute
	DefaultSchemaToDump                = "public"
	DefaultCreateMigrationSQLSeparator = "-- migrate:down"
)

type Config struct {
	*RootConfig

	Migration *MigrationConfig
	Create    *CreateConfig
}

func NewConfig() *Config {
	return &Config{
		RootConfig: &RootConfig{
			SchemaVersionTable:   DefaultSchemaVersionTable,
			AmigoFolderPath:      DefaultAmigoFolder,
			MigrationFolder:      DefaultMigrationFolder,
			MigrationPackageName: DefaultMigrationPackageName,
			ShellPath:            DefaultShellPath,
			PGDumpPath:           DefaultPGDumpPath,
			SchemaOutPath:        DefaultSchemaOutPath,
			SchemaToDump:         DefaultSchemaToDump,
		},
		Migration: &MigrationConfig{
			Timeout: DefaultTimeout,
			Steps:   1,
		},
		Create: &CreateConfig{
			Type:         string(types.MigrationFileTypeClassic),
			Dump:         false,
			SQLSeparator: DefaultCreateMigrationSQLSeparator,
			Skip:         false,
		},
	}
}

type RootConfig struct {
	CurrentContext            string
	AmigoFolderPath           string
	DSN                       string
	JSON                      bool
	ShowSQL                   bool
	ShowSQLSyntaxHighlighting bool
	MigrationFolder           string
	MigrationPackageName      string
	SchemaVersionTable        string
	ShellPath                 string
	PGDumpPath                string
	SchemaOutPath             string
	SchemaToDump              string
	Debug                     bool
}

func (r *Config) GetRealDSN() string {
	switch types.GetDriver(r.RootConfig.DSN) {
	case types.DriverSQLite:
		return strings.TrimPrefix(r.RootConfig.DSN, "sqlite:")
	}

	return r.RootConfig.DSN
}

func (r *RootConfig) ValidateDSN() error {
	if r.DSN == "" {
		return ErrDSNEmpty
	}

	return nil
}

type MigrationConfig struct {
	Version         string
	Steps           int
	DryRun          bool
	ContinueOnError bool
	Timeout         time.Duration
	UseSchemaDump   bool
	DumpSchemaAfter bool
}

func (m *MigrationConfig) ValidateVersion() error {
	if m.Version == "" {
		return nil
	}

	re := regexp.MustCompile(`\d{14}(_\w+)?\.(go|sql)`)
	if !re.MatchString(m.Version) {
		return fmt.Errorf("version must be in the format: 20240502083700 or 20240502083700_name.go or 20240502083700_name.sql")
	}

	return nil
}

type CreateConfig struct {
	Type string
	Dump bool

	SQLSeparator string

	Skip bool

	// Version is post setted after the name have been generated from the arg and time
	Version string
}

func (c *CreateConfig) ValidateType() error {
	allowedTypes := []string{string(types.MigrationFileTypeClassic), string(types.MigrationFileTypeChange), string(types.MigrationFileTypeSQL)}

	for _, t := range allowedTypes {
		if c.Type == t {
			return nil
		}
	}

	return fmt.Errorf("unsupported type, allowed types are: %s", strings.Join(allowedTypes, ", "))
}

func MergeConfig(toMerge Config) *Config {
	defaultCtx := NewConfig()

	if toMerge.RootConfig != nil {
		if toMerge.RootConfig.AmigoFolderPath != "" {
			defaultCtx.RootConfig.AmigoFolderPath = toMerge.RootConfig.AmigoFolderPath
		}

		if toMerge.RootConfig.DSN != "" {
			defaultCtx.RootConfig.DSN = toMerge.RootConfig.DSN
		}

		if toMerge.RootConfig.JSON {
			defaultCtx.RootConfig.JSON = toMerge.RootConfig.JSON
		}

		if toMerge.RootConfig.ShowSQL {
			defaultCtx.RootConfig.ShowSQL = toMerge.RootConfig.ShowSQL
		}

		if toMerge.RootConfig.MigrationFolder != "" {
			defaultCtx.RootConfig.MigrationFolder = toMerge.RootConfig.MigrationFolder
		}

		if toMerge.RootConfig.MigrationPackageName != "" {
			defaultCtx.RootConfig.MigrationPackageName = toMerge.RootConfig.MigrationPackageName
		}

		if toMerge.RootConfig.SchemaVersionTable != "" {
			defaultCtx.RootConfig.SchemaVersionTable = toMerge.RootConfig.SchemaVersionTable
		}

		if toMerge.RootConfig.ShellPath != "" {
			defaultCtx.RootConfig.ShellPath = toMerge.RootConfig.ShellPath
		}

		if toMerge.RootConfig.PGDumpPath != "" {
			defaultCtx.RootConfig.PGDumpPath = toMerge.RootConfig.PGDumpPath
		}

		if toMerge.RootConfig.Debug {
			defaultCtx.RootConfig.Debug = toMerge.RootConfig.Debug
		}

		if toMerge.RootConfig.SchemaToDump != "" {
			defaultCtx.RootConfig.SchemaToDump = toMerge.RootConfig.SchemaToDump
		}

		if toMerge.RootConfig.SchemaOutPath != "" {
			defaultCtx.RootConfig.SchemaOutPath = toMerge.RootConfig.SchemaOutPath
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

// WithAmigoFolder sets the folder where amigo files are stored
// DefaultYamlConfig is "db"
func (a *Config) WithAmigoFolder(folder string) *Config {
	a.RootConfig.AmigoFolderPath = folder
	return a
}

// WithMigrationFolder sets the folder where migration files are stored
// DefaultYamlConfig is "db/migrations"
func (a *Config) WithMigrationFolder(folder string) *Config {
	a.RootConfig.MigrationFolder = folder
	return a
}

// WithMigrationPackageName sets the package name where migration files are stored
// DefaultYamlConfig is "migrations"
func (a *Config) WithMigrationPackageName(packageName string) *Config {
	a.RootConfig.MigrationPackageName = packageName
	return a
}

// WithSchemaVersionTable sets the table name where the schema version is stored
// DefaultYamlConfig is "public.mig_schema_versions"
func (a *Config) WithSchemaVersionTable(table string) *Config {
	a.RootConfig.SchemaVersionTable = table
	return a
}

// WithDSN sets the DSN to use
// To use SQLite, use "sqlite:path/to/file.db"
func (a *Config) WithDSN(dsn string) *Config {
	a.RootConfig.DSN = dsn
	return a
}

// WithSchemaOutPath sets if the schema should be dumped before migration
// DefaultYamlConfig is "db/schema.sql"
func (a *Config) WithSchemaOutPath(path string) *Config {
	a.RootConfig.SchemaOutPath = path
	return a
}

// WithMigrationDumpSchemaAfterMigrating sets if the schema should be dumped after migration
func (a *Config) WithMigrationDumpSchemaAfterMigrating(dumpSchema bool) *Config {
	if a.Migration == nil {
		a.Migration = &MigrationConfig{}
	}
	a.Migration.DumpSchemaAfter = dumpSchema

	return a
}

// WithShowSQL sets if the SQL should be shown in the output
func (a *Config) WithShowSQL(showSQL bool) *Config {
	a.RootConfig.ShowSQL = showSQL
	return a
}

// WithJSON sets if the output should be in JSON
func (a *Config) WithJSON(json bool) *Config {
	a.RootConfig.JSON = json
	return a
}

// WithShowSQLSyntaxHighlighting sets if the SQL should be highlighted
func (a *Config) WithShowSQLSyntaxHighlighting(highlight bool) *Config {
	a.RootConfig.ShowSQLSyntaxHighlighting = highlight
	return a
}

// WithShellPath sets the path to the shell (used to execute commands like pg_dump)
func (a *Config) WithShellPath(path string) *Config {
	a.RootConfig.ShellPath = path
	return a
}

// WithPGDumpPath sets the path to the pg_dump executable
// default is "pg_dump"
func (a *Config) WithPGDumpPath(path string) *Config {
	a.RootConfig.PGDumpPath = path
	return a
}

// WithSchemaToDump sets the schema to dump
// default is "public"
func (a *Config) WithSchemaToDump(schema string) *Config {
	a.RootConfig.SchemaToDump = schema
	return a
}

// WithDebug sets if the debug mode should be enabled
func (a *Config) WithDebug(debug bool) *Config {
	a.RootConfig.Debug = debug
	return a
}

// WithMigrationDryRun sets if the migration should be dry run
// when true, the migration will not be executed, it's wrapped in a transaction
func (a *Config) WithMigrationDryRun(dryRun bool) *Config {
	if a.Migration == nil {
		a.Migration = &MigrationConfig{}
	}
	a.Migration.DryRun = dryRun
	return a
}

// WithMigrationContinueOnError sets if the migration should continue on error
func (a *Config) WithMigrationContinueOnError(continueOnError bool) *Config {
	if a.Migration == nil {
		a.Migration = &MigrationConfig{}
	}
	a.Migration.ContinueOnError = continueOnError
	return a
}

// WithMigrationTimeout sets the timeout for the migration
func (a *Config) WithMigrationTimeout(timeout time.Duration) *Config {
	if a.Migration == nil {
		a.Migration = &MigrationConfig{}
	}
	a.Migration.Timeout = timeout
	return a
}

// WithMigrationVersion sets the version of the migration
// format be like: 20240502083700 or 20240502083700_name.{go, sql}
func (a *Config) WithMigrationVersion(version string) *Config {
	if a.Migration == nil {
		a.Migration = &MigrationConfig{}
	}
	a.Migration.Version = version
	return a
}

// WithMigrationSteps sets the number of steps to migrate
// useful for rolling back
func (a *Config) WithMigrationSteps(steps int) *Config {
	if a.Migration == nil {
		a.Migration = &MigrationConfig{}
	}
	a.Migration.Steps = steps
	return a
}

// WithMigrationUseSchemaDump sets if the schema should be dumped before migration
// when true, if there is no migrations and a schema dump exists, it will be used instead of applying all migrations
func (a *Config) WithMigrationUseSchemaDump(useSchemaDump bool) *Config {
	if a.Migration == nil {
		a.Migration = &MigrationConfig{}
	}
	a.Migration.UseSchemaDump = useSchemaDump
	return a
}

// WithMigrationDumpSchemaAfter sets if the schema should be dumped after migration
func (a *Config) WithMigrationDumpSchemaAfter(dumpSchemaAfter bool) *Config {
	if a.Migration == nil {
		a.Migration = &MigrationConfig{}
	}
	a.Migration.DumpSchemaAfter = dumpSchemaAfter
	return a
}

// WithCreateType sets the type of the migration file
func (a *Config) WithCreateType(createType types.MigrationFileType) *Config {
	if a.Create == nil {
		a.Create = &CreateConfig{}
	}
	a.Create.Type = string(createType)
	return a
}

// WithCreateDump sets if the created file should contains the dump of the database
func (a *Config) WithCreateDump(dump bool) *Config {
	if a.Create == nil {
		a.Create = &CreateConfig{}
	}
	a.Create.Dump = dump
	return a
}

// WithCreateSQLSeparator sets the separator to split the  down part of the migration in type sql
// DefaultYamlConfig value is "-- migrate:down"
func (a *Config) WithCreateSQLSeparator(separator string) *Config {
	if a.Create == nil {
		a.Create = &CreateConfig{}
	}
	a.Create.SQLSeparator = separator
	return a
}

func (a *Config) WithCreateSkip(skip bool) *Config {
	if a.Create == nil {
		a.Create = &CreateConfig{}
	}
	a.Create.Skip = skip
	return a
}

func (a *Config) WithCreateVersion(version string) *Config {
	if a.Create == nil {
		a.Create = &CreateConfig{}
	}
	a.Create.Version = version
	return a
}
