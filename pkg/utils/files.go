package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CreateOrOpenFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		if !os.IsExist(err) {
			return nil, fmt.Errorf("unable to create parent directory: %w", err)
		}
	}

	// create or open file
	return os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
}

func GetFileContent(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func EnsurePrentDirExists(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("unable to create parent directory: %w", err)
		}
	}

	return nil
}
