package templates

type (
	MigrationsData struct {
		Package    string
		Migrations []string
	}

	MigrationData struct {
		Package    string
		StructName string
		Name       string

		InUp   string
		InDown string

		CreatedAt string // RFC3339

		PackageDriverName string
		PackageDriverPath string

		UseSchemaImport bool
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
