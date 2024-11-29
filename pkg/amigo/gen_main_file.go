package amigo

import (
	"fmt"
	"io"
	"path"

	"github.com/alexisvisco/amigo/pkg/templates"
	"github.com/alexisvisco/amigo/pkg/utils"
)

func (a Amigo) GenerateMainFile(writer io.Writer) error {
	var (
		migrationFolder = a.Config.MigrationFolder
	)

	name, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("unable to get module name: %w", err)
	}

	packagePath := path.Join(name, migrationFolder)

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
