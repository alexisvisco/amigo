package amigo

import (
	"fmt"
	"io"
	"path"

	"github.com/alexisvisco/amigo/pkg/templates"
	"github.com/alexisvisco/amigo/pkg/utils"
)

// GenerateMainFile generate the main.go file in the amigo folder
func (a Amigo) GenerateMainFile(writer io.Writer) error {
	name, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("unable to get module name: %w", err)
	}

	packagePath := path.Join(name, a.ctx.MigrationFolder)

	template, err := templates.GetMainTemplate(templates.MainData{
		PackagePath: packagePath,
		DriverPath:  a.Driver.PackagePath(),
		DriverName:  a.Driver.String(),
	})

	if err != nil {
		return fmt.Errorf("unable to get main template: %w", err)
	}

	_, err = writer.Write([]byte(template))
	if err != nil {
		return fmt.Errorf("unable to write main file: %w", err)
	}

	return nil
}
