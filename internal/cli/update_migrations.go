package cli

import (
	"fmt"
	"github.com/alexisvisco/mig/internal/templates"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

func GenerateMigrationsFile(folder, pkg, filePath string) error {
	// Update file content
	updatedContent, err := updateMigrationsFileSlice(folder, pkg)
	if err != nil {
		return fmt.Errorf("unable to update file content: %w", err)
	}

	// Write file
	err = os.WriteFile(filePath, []byte(updatedContent), 0644)

	return nil
}

// updateMigrationsFileSlice updates the migrations file content by adding the structToAdd element
// into the migration slice in the content. The function returns the updated content.
func updateMigrationsFileSlice(folder, pkg string) (string, error) {
	// get all files that start with a number
	re := regexp.MustCompile(`(\d+)_(.*)\.go`)

	migrationFiles := make(map[int64]string)

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if re.MatchString(info.Name()) {
				matches := re.FindStringSubmatch(info.Name())
				fileNumber := matches[1]
				migrationName := matches[2]

				atoi, _ := strconv.ParseInt(fileNumber, 10, 64)
				migrationFiles[atoi] = migrationName
			}
		}

		return nil
	})

	// sort the files
	var keys []int64
	for k := range migrationFiles {
		keys = append(keys, k)
	}

	// sort the keys in ascending order
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	var migrations []string
	for _, k := range keys {
		fmt.Println(migrationFiles[k], k)
		migrations = append(migrations, MigrationStructName(migrationFiles[k], k))
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
