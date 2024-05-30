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
	DriverUnknown  Driver = ""
	DriverPostgres Driver = "postgres"
)

var DriverValues = []Driver{
	DriverPostgres,
}

func (d Driver) PackageSchemaPath() string {
	switch d {
	case DriverPostgres:
		return "github.com/alexisvisco/amigo/pkg/schema/pg"
	default:
		return "github.com/alexisvisco/amigo/pkg/schema/base"
	}
}

func (d Driver) PackagePath() string {
	switch d {
	case DriverPostgres:
		return "github.com/jackc/pgx/v5/stdlib"
	default:
		return "your_driver_here"
	}
}

func (d Driver) PackageName() string {
	return filepath.Base(d.PackageSchemaPath())
}

func (d Driver) String() string {
	switch d {
	case DriverPostgres:
		return "pgx"
	default:
		return "your_driver_here"
	}
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
