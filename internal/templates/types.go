package templates

type (
	MigrationsData struct {
		Package    string
		Migrations []string
	}

	MigrationData struct {
		Package    string
		StructName string
		Driver     string
		Name       string

		InUp   string
		InDown string

		CreatedAt string // RFC3339
	}
)
