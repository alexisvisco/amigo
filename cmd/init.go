package cmd

import (
	"fmt"
	"path"
	"time"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/templates"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"gopkg.in/yaml.v3"
)

func executeInit(
	mainFilePath,
	amigoFolder,
	table,
	migrationsFolder string,
) error {
	// create the main file
	logger.Info(events.FolderAddedEvent{FolderName: amigoFolder})

	file, err := utils.CreateOrOpenFile(mainFilePath)
	if err != nil {
		return fmt.Errorf("unable to open main.go file: %w", err)
	}

	cfg := amigoconfig.NewConfig().
		WithAmigoFolder(amigoFolder).
		WithMigrationFolder(migrationsFolder).
		WithSchemaVersionTable(table)

	am := amigo.NewAmigo(cfg)

	err = am.GenerateMainFile(file)
	if err != nil {
		return err
	}

	logger.Info(events.FileAddedEvent{FileName: mainFilePath})

	// create the base schema version table
	now := time.Now()
	migrationFileName := fmt.Sprintf("%s_create_table_schema_version.go", now.UTC().Format(utils.FormatTime))
	file, err = utils.CreateOrOpenFile(path.Join(cfg.MigrationFolder, migrationFileName))
	if err != nil {
		return fmt.Errorf("unable to open migrationsFolder.go file: %w", err)
	}

	inUp, err := templates.GetInitCreateTableTemplate(templates.CreateTableData{Name: table},
		am.Driver == types.DriverUnknown)
	if err != nil {
		return err
	}

	err = am.GenerateMigrationFile(&amigo.GenerateMigrationFileParams{
		Name:            "create_table_schema_version",
		Up:              inUp,
		Down:            "// nothing to do to keep the schema version table",
		Type:            types.MigrationFileTypeClassic,
		Now:             now,
		Writer:          file,
		UseSchemaImport: am.Driver != types.DriverUnknown,
		UseFmtImport:    am.Driver == types.DriverUnknown,
	})
	if err != nil {
		return err
	}
	logger.Info(events.FileAddedEvent{FileName: path.Join(migrationsFolder, migrationFileName)})

	// create the migrationsFolder file where all the migrationsFolder will be stored
	file, err = utils.CreateOrOpenFile(path.Join(amigoFolder, "migrations.go"))
	if err != nil {
		return err
	}

	err = am.GenerateMigrationsFiles(file)
	if err != nil {
		return err
	}

	logger.Info(events.FileAddedEvent{FileName: path.Join(amigoFolder, migrationFileName)})

	// write the context file
	out, err := yaml.Marshal(amigoconfig.DefaultYamlConfig)
	if err != nil {
		return err
	}

	openFile, err := utils.CreateOrOpenFile(path.Join(amigoFolder, amigoconfig.FileName))
	if err != nil {
		return fmt.Errorf("unable to open config file: %w", err)
	}
	defer openFile.Close()

	_, err = openFile.WriteString(string(out))
	if err != nil {
		return fmt.Errorf("unable to write config file: %w", err)
	}

	return nil
}
