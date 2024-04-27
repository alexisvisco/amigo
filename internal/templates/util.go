package templates

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed migrations.go.tmpl
var migrations string

//go:embed migration_change.go.tmpl
var migrationChange string

//go:embed migration_classic.go.tmpl
var migrationClassic string

func GetMigrationsTemplate(t MigrationsData) (string, error) {
	parse, err := template.New("migrations").Parse(migrations)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := parse.Execute(&tpl, t); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func GetMigrationChangeTemplate(migType string, t MigrationData) (string, error) {
	var tpl string
	switch migType {
	case "classic":
		tpl = migrationClassic
	case "change":
		tpl = migrationChange
	default:
		return "", nil
	}

	parse, err := template.New("migration").Parse(tpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := parse.Execute(&buf, t); err != nil {
		return "", err
	}

	return buf.String(), nil
}
