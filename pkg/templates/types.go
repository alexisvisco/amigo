package templates

import "github.com/alexisvisco/amigo/pkg/types"

type (
	MigrationsData struct {
		Package    string
		Migrations []string
	}

	MigrationData struct {
		Package    string
		StructName string
		Name       string

		Type types.MigrationFileType

		Imports []string

		InChange string
		InUp     string
		InDown   string

		CreatedAt string // RFC3339

		PackageDriverName string
		PackageDriverPath string

		UseSchemaImport bool
		UseFmtImport    bool
	}

	CreateTableData struct {
		Name string
	}

	MainData struct {
		PackagePath string
		DriverPath  string
		DriverName  string
	}
)
