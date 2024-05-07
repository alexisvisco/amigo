package utils

import (
	"fmt"
	"golang.org/x/mod/modfile"
	"os"
)

func GetModuleName() (string, error) {
	goModBytes, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod file: %w", err)
	}

	modName := modfile.ModulePath(goModBytes)

	return modName, nil
}
