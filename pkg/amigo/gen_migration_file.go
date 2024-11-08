package amigo

import (
	"fmt"
	"io"
	"time"

	"github.com/alexisvisco/amigo/pkg/templates"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/gobuffalo/flect"
)

type GenerateMigrationFileParams struct {
	Name            string
	Up              string
	Down            string
	Change          string
	Type            types.MigrationFileType
	Now             time.Time
	UseSchemaImport bool
	UseFmtImport    bool
	Writer          io.Writer
}

// GenerateMigrationFile generate a migration file in the migrations folder
func (a Amigo) GenerateMigrationFile(params *GenerateMigrationFileParams) error {

	structName := utils.MigrationStructName(params.Now, params.Name)

	orDefault := func(s string) string {
		if s == "" {
			return "// TODO: implement the migration"
		}
		return s
	}

	fileContent, err := templates.GetMigrationTemplate(templates.MigrationData{
		IsSQL:             params.Type == types.MigrationFileTypeSQL,
		Package:           a.ctx.PackagePath,
		StructName:        structName,
		Name:              flect.Underscore(params.Name),
		Type:              params.Type,
		InChange:          orDefault(params.Change),
		InUp:              orDefault(params.Up),
		InDown:            orDefault(params.Down),
		CreatedAt:         params.Now.Format(time.RFC3339),
		PackageDriverName: a.Driver.PackageName(),
		PackageDriverPath: a.Driver.PackageSchemaPath(),
		UseSchemaImport:   params.UseSchemaImport,
		UseFmtImport:      params.UseFmtImport,
	})

	if err != nil {
		return fmt.Errorf("unable to get migration template: %w", err)
	}

	_, err = params.Writer.Write([]byte(fileContent))
	if err != nil {
		return fmt.Errorf("unable to write migration file: %w", err)
	}

	return nil
}
