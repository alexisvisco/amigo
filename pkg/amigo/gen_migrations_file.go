package amigo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/alexisvisco/amigo/pkg/templates"
	"github.com/alexisvisco/amigo/pkg/utils"
)

// GenerateMigrationsFiles generate the migrations file in the migrations folder
// It's used to keep track of all migrations
func (a Amigo) GenerateMigrationsFiles(writer io.Writer) error {
	migrationFiles, keys, err := a.getMigrationFiles(true)
	if err != nil {
		return err
	}

	var migrations []string
	var mustImportSchemaPackage *string
	for _, k := range keys {
		if migrationFiles[k].isSQL {
			// schema.NewSQLMigration[*pg.Schema](sqlMigrationsFS, "20240602081806_drop_index.sql", "2024-06-02T10:18:06+02:00", "---- down:"),
			line := fmt.Sprintf("schema.NewSQLMigration[%s](sqlMigrationsFS, \"%s\", \"%s\", \"%s\")",
				a.Driver.StructName(),
				migrationFiles[k].fulName,
				k.Format(time.RFC3339),
				a.Config.Create.SQLSeparator,
			)

			migrations = append(migrations, line)

			if mustImportSchemaPackage == nil {
				v := a.Driver.PackageSchemaPath()
				mustImportSchemaPackage = &v
			}
		} else {
			migrations = append(migrations, fmt.Sprintf("&%s{}", utils.MigrationStructName(k, migrationFiles[k].Name)))

		}
	}

	content, err := templates.GetMigrationsTemplate(templates.MigrationsData{
		Package:             a.Config.MigrationPackageName,
		Migrations:          migrations,
		ImportSchemaPackage: mustImportSchemaPackage,
	})

	if err != nil {
		return fmt.Errorf("unable to get migrations template: %w", err)
	}

	_, err = writer.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("unable to write migrations file: %w", err)
	}

	return nil
}

type migrationFile struct {
	Name    string
	fulName string
	isSQL   bool
}

func (a Amigo) getMigrationFiles(ascending bool) (map[time.Time]migrationFile, []time.Time, error) {
	migrationFiles := make(map[time.Time]migrationFile)

	// get the list of structs by the file name
	err := filepath.Walk(a.Config.MigrationFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if utils.MigrationFileRegexp.MatchString(info.Name()) {
				matches := utils.MigrationFileRegexp.FindStringSubmatch(info.Name())
				fileTime := matches[1]
				migrationName := matches[2]
				ext := matches[3]

				t, _ := time.Parse(utils.FormatTime, fileTime)
				migrationFiles[t] = migrationFile{Name: migrationName, isSQL: ext == "sql", fulName: info.Name()}
			}
		}

		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to walk through the migration folder: %w", err)
	}

	// sort the files
	var keys []time.Time
	for k := range migrationFiles {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		if ascending {
			return keys[i].Unix() < keys[j].Unix()
		} else {
			return keys[i].Unix() > keys[j].Unix()
		}
	})

	return migrationFiles, keys, nil
}
