package amigo

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// cliGenerate creates a new migration file
func (c *CLI) cliGenerate(args []string) int {
	// Show help if requested (before parsing to avoid flag.Parse handling it)
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		c.cliGenerateHelp()
		return 0
	}

	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(c.errorOutput)

	var format string
	fs.StringVar(&format, "format", c.defaultFileFormat, "File format (sql or go)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	// Validate format
	if format != "sql" && format != "go" {
		fmt.Fprintf(c.errorOutput, "%s\n", c.cliOutput.error(fmt.Sprintf("Error: invalid format '%s', must be 'sql' or 'go'", format)))
		return 1
	}

	// Get migration name from remaining args
	if fs.NArg() == 0 {
		fmt.Fprintln(c.errorOutput, "Error: migration name is required")
		fmt.Fprintln(c.errorOutput, "")
		c.cliGenerateHelp()
		return 1
	}

	name := strings.Join(fs.Args(), "_")

	// Generate timestamp (YYYYMMDDHHMMSS format in UTC)
	timestamp := time.Now().UTC().Format("20060102150405")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(c.directory, 0755); err != nil {
		fmt.Fprintf(c.errorOutput, "%s\n", c.cliOutput.error(fmt.Sprintf("Error: failed to create directory: %v", err)))
		return 1
	}

	var filename, content string
	var err error

	if format == "sql" {
		filename = fmt.Sprintf("%s_%s.sql", timestamp, name)
		content, err = c.generateSQLTemplate()
	} else {
		filename = fmt.Sprintf("%s_%s.go", timestamp, name)
		content, err = c.generateGoTemplate(name, timestamp)
	}

	if err != nil {
		fmt.Fprintf(c.errorOutput, "%s\n", c.cliOutput.error(fmt.Sprintf("Error: failed to generate template: %v", err)))
		return 1
	}

	filepath := filepath.Join(c.directory, filename)

	// Check if file already exists
	if _, err := os.Stat(filepath); err == nil {
		fmt.Fprintf(c.errorOutput, "%s\n", c.cliOutput.error(fmt.Sprintf("Error: file already exists: %s", filepath)))
		return 1
	}

	// Write file
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		fmt.Fprintf(c.errorOutput, "%s\n", c.cliOutput.error(fmt.Sprintf("Error: failed to write file: %v", err)))
		return 1
	}

	fmt.Fprintf(c.output, "Created migration: %s\n", c.cliOutput.path(filepath))

	// Regenerate migrations.go
	if err := c.generateMigrationsList(); err != nil {
		fmt.Fprintf(c.errorOutput, "%s\n", c.cliOutput.error(fmt.Sprintf("Error: failed to regenerate migrations.go: %v", err)))
		return 1
	}

	fmt.Fprintf(c.output, "Updated migrations list: %s\n", c.cliOutput.path(c.directory+"/migrations.go"))

	return 0
}

// generateSQLTemplate returns the SQL migration template
func (c *CLI) generateSQLTemplate() (string, error) {
	tmpl, err := template.New("sql").Parse(sqlTemplate)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{
		"UpAnnotation":   c.config.SQLFileUpAnnotation,
		"DownAnnotation": c.config.SQLFileDownAnnotation,
		"Transactional":  c.defaultTransactional,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateGoTemplate returns the Go migration template
func (c *CLI) generateGoTemplate(name, timestamp string) (string, error) {
	tmpl, err := template.New("go").Parse(goTemplate)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{
		"PackageName":   c.packageName,
		"StructName":    sanitizeName(name),
		"Name":          name,
		"Timestamp":     timestamp,
		"Transactional": c.defaultTransactional,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// sanitizeName converts a migration name to a valid Go identifier
func sanitizeName(name string) string {
	// Simple title case conversion
	parts := strings.Split(name, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// cliGenerateHelp displays help for the generate command
func (c *CLI) cliGenerateHelp() {
	help := `Usage: generate [options] <name>

Generate a new migration file with the given name.

Options:
  --format string    File format: 'sql' or 'go' (default: configured value)
  -h, --help         Show this help message

Arguments:
  name               Name of the migration (can contain spaces or underscores)

Examples:
  generate create_users_table
  generate --format=go add_email_column
  generate "create users table"
`
	fmt.Fprint(c.output, help)
}
