package mig

import (
	"fmt"
	"github.com/alexisvisco/mig/pkg/templates"
	"github.com/alexisvisco/mig/pkg/utils"
	"os"
	"path"
)

func GenerateMainFile(folder, migrationFolder string) error {
	name, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("unable to get module name: %w", err)
	}

	packagePath := path.Join(name, migrationFolder)

	template, err := templates.GetMainTemplate(templates.MainData{
		PackagePath: packagePath,
	})

	if err != nil {
		return fmt.Errorf("unable to get main template: %w", err)
	}

	err = os.WriteFile(path.Join(folder, "main.go"), []byte(template), 0644)
	if err != nil {
		return fmt.Errorf("unable to write main file: %w", err)
	}

	return nil
}
