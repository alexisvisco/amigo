package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsDirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
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

func HasFilesWithExtension(folder string, ext ...string) bool {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return false
	}
	for _, file := range files {
		for _, e := range ext {
			if strings.HasSuffix(file.Name(), e) {
				return true
			}
		}
		return false
	}
	return false
}
