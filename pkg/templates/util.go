package templates

import (
	"bytes"
	_ "embed"
	"github.com/alexisvisco/mig/pkg/types"
	"text/template"
)

//go:embed migrations.go.tmpl
var migrationsList string

//go:embed migration_change.go.tmpl
var migrationChange string

//go:embed migration_classic.go.tmpl
var migrationClassic string

//go:embed init_create_table.go.tmpl
var initCreateTable string

//go:embed main.go.tmpl
var main string

func GetMigrationsTemplate(t MigrationsData) (string, error) {
	parse, err := template.New("migrationsList").Parse(migrationsList)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := parse.Execute(&tpl, t); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func GetMigrationChangeTemplate(direction types.MigrationFileType, t MigrationData) (string, error) {
	var tpl string
	switch direction {
	case types.MigrationFileTypeClassic:
		tpl = migrationClassic
	case types.MigrationFileTypeChange:
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

func GetInitCreateTableTemplate(t CreateTableData) (string, error) {
	parse, err := template.New("initCreateTable").Parse(initCreateTable)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := parse.Execute(&tpl, t); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func GetMainTemplate(t MainData) (string, error) {
	parse, err := template.New("main").Parse(main)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := parse.Execute(&tpl, t); err != nil {
		return "", err
	}

	return tpl.String(), nil
}
