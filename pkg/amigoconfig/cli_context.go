package amigoconfig

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

var FileName = "config.yml"

type YamlConfig struct {
	ShellPath                 string `yaml:"shell-path"`
	Debug                     bool   `yaml:"debug"`
	ShowSQL                   bool   `yaml:"show-sql"`
	ShowSQLSyntaxHighlighting bool   `yaml:"show-sql-syntax-highlighting"`

	CurrentContext string                       `yaml:"current-context"`
	Contexts       map[string]YamlConfigContext `yaml:"contexts"`
}

type YamlConfigContext struct {
	SchemaVersionTable   string        `yaml:"schema-version-table"`
	MigrationFolder      string        `yaml:"migration-folder"`
	MigrationPackageName string        `yaml:"migration-package-name"`
	SchemaToDump         string        `yaml:"schema-to-dump"`
	SchemaOutPath        string        `yaml:"schema-out-path"`
	Timeout              time.Duration `yaml:"timeout"`
	DSN                  string        `yaml:"dsn"`

	PGDumpPath string `yaml:"pg-dump-path"`
}

var DefaultYamlConfig = YamlConfig{
	ShellPath:                 DefaultShellPath,
	Debug:                     false,
	ShowSQL:                   true,
	ShowSQLSyntaxHighlighting: true,

	CurrentContext: "default",
	Contexts: map[string]YamlConfigContext{
		"default": {
			SchemaVersionTable:   DefaultSchemaVersionTable,
			MigrationFolder:      DefaultMigrationFolder,
			MigrationPackageName: DefaultMigrationPackageName,
			SchemaToDump:         DefaultSchemaToDump,
			SchemaOutPath:        DefaultSchemaOutPath,
			Timeout:              DefaultTimeout,
			DSN:                  "postgres://user:password@host:port/dbname?sslmode=disable",
			PGDumpPath:           DefaultPGDumpPath,
		},
	},
}

func LoadYamlConfig(path string) (*YamlConfig, error) {
	var config YamlConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read contexts file: %w", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) OverrideWithYamlConfig(yaml *YamlConfig) {
	if yaml == nil || len(yaml.Contexts) == 0 {
		return
	}

	context, ok := yaml.Contexts[yaml.CurrentContext]
	if !ok {
		return
	}

	// Override root config values
	if yaml.ShellPath != "" {
		c.RootConfig.ShellPath = yaml.ShellPath
	}

	if yaml.Debug {
		c.RootConfig.Debug = yaml.Debug
	}

	if yaml.ShowSQL {
		c.RootConfig.ShowSQL = yaml.ShowSQL
	}

	if yaml.ShowSQLSyntaxHighlighting {
		c.RootConfig.ShowSQLSyntaxHighlighting = yaml.ShowSQLSyntaxHighlighting
	}

	if yaml.CurrentContext != "" && c.RootConfig.CurrentContext == "" {
		c.RootConfig.CurrentContext = yaml.CurrentContext
	}

	// Override per-driver config values
	if context.SchemaVersionTable != "" {
		c.RootConfig.SchemaVersionTable = context.SchemaVersionTable
	}
	if context.MigrationFolder != "" {
		c.RootConfig.MigrationFolder = context.MigrationFolder
	}
	if context.MigrationPackageName != "" {
		c.RootConfig.MigrationPackageName = context.MigrationPackageName
	}
	if context.SchemaToDump != "" {
		c.RootConfig.SchemaToDump = context.SchemaToDump
	}
	if context.SchemaOutPath != "" {
		c.RootConfig.SchemaOutPath = context.SchemaOutPath
	}
	if context.PGDumpPath != "" {
		c.RootConfig.PGDumpPath = context.PGDumpPath
	}
	if context.Timeout != 0 {
		if c.Migration == nil {
			c.Migration = &MigrationConfig{}
		}
		c.Migration.Timeout = context.Timeout
	}
	if context.DSN != "" {
		c.RootConfig.DSN = context.DSN
	}

	return
}
