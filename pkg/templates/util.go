package templates

import (
	"bytes"
	_ "embed"
	"sort"
	"strings"
	"text/template"
)

//go:embed migrations.go.tmpl
var migrationsList string

//go:embed migration.go.tmpl
var migration string

//go:embed migration.sql.tmpl
var migrationSQL string

//go:embed init_create_table.go.tmpl
var initCreateTable string

//go:embed init_create_table_base.go.tmpl
var initCreateTableBase string

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

func GetMigrationTemplate(t MigrationData) (string, error) {
	if t.IsSQL {
		return migrationSQL, nil
	}

	t.Imports = append(t.Imports, "time")
	t.Imports = append(t.Imports, "github.com/alexisvisco/amigo/pkg/schema/"+t.PackageDriverName)

	if t.UseSchemaImport {
		t.Imports = append(t.Imports, "github.com/alexisvisco/amigo/pkg/schema")
	}

	if t.UseFmtImport {
		t.Imports = append(t.Imports, "fmt")
	}

	sort.Strings(t.Imports)

	parse, err := template.New("migration").Funcs(funcMap).Parse(migration)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := parse.Execute(&buf, t); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func GetInitCreateTableTemplate(t CreateTableData, base bool) (string, error) {

	tmpl := initCreateTable
	if base {
		tmpl = initCreateTableBase
	}
	parse, err := template.New("initCreateTable").Parse(tmpl)
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
	parse, err := template.New("main").Funcs(funcMap).Parse(main)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := parse.Execute(&tpl, t); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

var funcMap = template.FuncMap{
	// indent the string with n tabs
	"indent": func(n int, s string) string {
		return strings.ReplaceAll(s, "\n", "\n"+strings.Repeat("\t", n))
	},
}
