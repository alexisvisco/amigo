package amigo

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/templates"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/gobuffalo/flect"
	"os"
	"path"
	"time"
)

type GenerateMigrationFileOptions struct {
	Name    string
	Folder  string
	Driver  types.Driver
	Package string
	MigType types.MigrationFileType
	InUp    string
	InDown  string
	Now     time.Time
}

func GenerateMigrationFile(opts GenerateMigrationFileOptions) (fileCreated string, version int64, err error) {
	now := time.Now()
	if !opts.Now.IsZero() {
		now = opts.Now
	}

	structName := utils.MigrationStructName(now, opts.Name)
	fileCreated = path.Join(opts.Folder, utils.MigrationFileFormat(now, opts.Name))
	fileContent, err := templates.GetMigrationChangeTemplate(opts.MigType, templates.MigrationData{
		Package:           opts.Package,
		PackageDriverName: opts.Driver.PackageName(),
		PackageDriverPath: opts.Driver.PackagePath(),
		StructName:        structName,
		Name:              flect.Underscore(opts.Name),
		InUp:              opts.InUp,
		InDown:            opts.InDown,
		CreatedAt:         now.Format(time.RFC3339),
	})
	if err != nil {
		return "", 0, fmt.Errorf("unable to get migration template: %w", err)
	}

	err = os.WriteFile(fileCreated, []byte(fileContent), 0644)
	if err != nil {
		return "", 0, fmt.Errorf("unable to write migration file %s: %w", fileCreated, err)
	}

	return fileCreated, now.Unix(), nil
}
