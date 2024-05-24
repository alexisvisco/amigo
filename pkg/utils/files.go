package utils

import (
	"fmt"
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
