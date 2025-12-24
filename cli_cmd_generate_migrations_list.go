package amigo

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

const migrationsListTemplate = `package migrations

import (
	"embed"

	"github.com/alexisvisco/amigo"
)

//go:embed *.sql
var sqlFiles embed.FS

func Migrations(cfg amigo.Configuration) []amigo.Migration {
	return []amigo.Migration{
{{range .SQLMigrations}}		amigo.SQLFileToMigration(sqlFiles, "{{.}}", cfg),
{{end}}{{range .GoMigrations}}		&Migration{{.}}{},
{{end}}	}
}
`

// generateMigrationsList creates the migrations.go file with all migrations
func (c *CLI) generateMigrationsList() error {
	// Read directory
	entries, err := os.ReadDir(c.config.Directory)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var sqlFiles []string
	var goStructNames []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip migrations.go itself
		if name == "migrations.go" {
			continue
		}

		ext := filepath.Ext(name)

		if ext == ".sql" {
			sqlFiles = append(sqlFiles, name)
		} else if ext == ".go" {
			// Extract struct name from filename
			// Format: YYYYMMDDHHMMSS_name.go -> Migration20231224120000Name
			baseName := strings.TrimSuffix(name, ".go")
			parts := strings.SplitN(baseName, "_", 2)
			if len(parts) == 2 {
				timestamp := parts[0]
				structName := sanitizeName(parts[1])
				fullStructName := timestamp + structName
				goStructNames = append(goStructNames, fullStructName)
			}
		}
	}

	// Sort for consistent output
	sort.Strings(sqlFiles)
	sort.Strings(goStructNames)

	// Generate migrations.go content
	tmpl, err := template.New("migrations_list").Parse(migrationsListTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	data := map[string]interface{}{
		"SQLMigrations": sqlFiles,
		"GoMigrations":  goStructNames,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Write migrations.go
	migrationsFile := filepath.Join(c.config.Directory, "migrations.go")
	if err := os.WriteFile(migrationsFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write migrations.go: %w", err)
	}

	return nil
}
