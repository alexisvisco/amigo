package types

import (
	"path/filepath"
	"strings"
)

type MigrationDirection string

const (
	MigrationDirectionUp   MigrationDirection = "UP"
	MigrationDirectionDown MigrationDirection = "DOWN"

	// MigrationDirectionNotReversible is used to indicate that the migration is reversed or is in a down type.
	// This is used to avoid infinite loop when executing a migration.
	// This is not a real migration direction. DO NOT ADD IT TO MigrationDirectionValues.
	MigrationDirectionNotReversible MigrationDirection = "NOT_REVERSIBLE"
)

var MigrationDirectionValues = []MigrationDirection{
	MigrationDirectionUp,
	MigrationDirectionDown,
}

func (m MigrationDirection) String() string {
	return strings.ToLower(string(m))
}

func (m MigrationDirection) IsValid() bool {
	for _, v := range MigrationDirectionValues {
		if v == m {
			return true
		}
	}

	return false
}

type Driver string

const (
	DriverPostgres Driver = "postgres"
)

var DriverValues = []Driver{
	DriverPostgres,
}

func (d Driver) PackagePath() string {
	switch d {
	case DriverPostgres:
		return "github.com/alexisvisco/mig/pkg/schema/pg"
	}

	return ""
}

func (d Driver) PackageName() string {
	return filepath.Base(d.PackagePath())
}

func (d Driver) String() string {
	return string(d)
}

func (d Driver) IsValid() bool {
	for _, v := range DriverValues {
		if v == d {
			return true
		}
	}

	return false
}

type MigrationFileType string

const (
	MigrationFileTypeChange  MigrationFileType = "change"
	MigrationFileTypeClassic MigrationFileType = "classic"
)

var MigrationFileTypeValues = []MigrationFileType{
	MigrationFileTypeChange,
	MigrationFileTypeClassic,
}

func (m MigrationFileType) String() string {
	return string(m)
}

func (m MigrationFileType) IsValid() bool {
	for _, v := range MigrationFileTypeValues {
		if v == m {
			return true
		}
	}

	return false
}

type MIGConfig struct {
	RootDSN            string `yaml:"root_dsn"`
	JSON               bool   `yaml:"json"`
	MigrationFolder    string `yaml:"migration_folder"`
	Package            string `yaml:"package"`
	SchemaVersionTable string `yaml:"schema_version_table"`
	ShellPath          string `yaml:"shell_path"`
	Verbose            bool   `yaml:"verbose"`
	MIGFolderPath      string `yaml:"mig_folder"`
}
