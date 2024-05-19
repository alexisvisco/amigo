package amigo

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/templates"
	"github.com/alexisvisco/amigo/pkg/utils"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func GenerateMigrationsFile(folder, pkg, filePath string) error {
	// Update file content
	updatedContent, err := updateMigrationsFileSlice(folder, pkg)
	if err != nil {
		return fmt.Errorf("unable to update file content: %w", err)
	}

	// Write file
	err = os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("unable to write file %s: %w", filePath, err)
	}

	return nil
}

// updateMigrationsFileSlice updates the migrations file content by adding the structToAdd element
// into the migration slice in the content. The function returns the updated content.
func updateMigrationsFileSlice(folder, pkg string) (string, error) {
	migrationFiles := make(map[time.Time]string)

	// get the list of structs by the file name
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if utils.FileRegexp.MatchString(info.Name()) {
				matches := utils.FileRegexp.FindStringSubmatch(info.Name())
				fileTime := matches[1]
				migrationName := matches[2]

				t, _ := time.Parse(utils.FormatTime, fileTime)
				migrationFiles[t] = migrationName
			}
		}

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("unable to walk through the folder: %w", err)
	}

	// sort the files
	var keys []time.Time
	for k := range migrationFiles {
		keys = append(keys, k)
	}

	// sort the keys in ascending order
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Unix() < keys[j].Unix()
	})

	var migrations []string
	for _, k := range keys {
		migrations = append(migrations, utils.MigrationStructName(k, migrationFiles[k]))
	}

	content, err := templates.GetMigrationsTemplate(templates.MigrationsData{
		Package:    pkg,
		Migrations: migrations,
	})

	if err != nil {
		return "", fmt.Errorf("unable to get migration template: %w", err)
	}

	return content, nil
}
