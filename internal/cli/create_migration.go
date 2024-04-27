package cli

import (
	"fmt"
	"github.com/alexisvisco/mig/internal/templates"
	"github.com/gobuffalo/flect"
	"os"
	"path"
	"time"
)

type CreateMigrationFileOptions struct {
	Name    string
	Folder  string
	Driver  string
	Package string
	MigType string
	InUp    string
	InDown  string
}

func MigrationStructName(name string, unix int64) string {
	return fmt.Sprintf("Migration%d%s", unix, flect.Pascalize(name))
}

func CreateMigrationFile(opts CreateMigrationFileOptions) (fileCreated, structName string, err error) {
	now := time.Now()
	structName = MigrationStructName(opts.Name, now.Unix())
	fileCreated = path.Join(opts.Folder, fmt.Sprintf("%d_%s.go", time.Now().Unix(), flect.Underscore(opts.Name)))
	fileContent, err := templates.GetMigrationChangeTemplate(opts.MigType, templates.MigrationData{
		Package:    opts.Package,
		StructName: structName,
		Driver:     flect.Pascalize(opts.Driver),
		Name:       flect.Underscore(opts.Name),
		InUp:       opts.InUp,
		InDown:     opts.InDown,
		CreatedAt:  now.Format(time.RFC3339),
	})
	if err != nil {
		return "", "", fmt.Errorf("unable to get migration template: %w", err)
	}

	err = os.WriteFile(fileCreated, []byte(fileContent), 0644)
	if err != nil {
		return "", "", fmt.Errorf("unable to write migration file %s: %w", fileCreated, err)
	}

	return
}
