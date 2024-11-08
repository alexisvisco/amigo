package schema

import (
	"errors"
	"fmt"
	"os"

	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
)

// tryMigrateWithSchemaDump tries to migrate with schema dump.
// this might be executed when the user arrives on a repo with a schema.sql, instead of running
// all the migrations we will try to dump the schema and apply it. Then tell we applied all versions.
func (m *Migrator[T]) tryMigrateWithSchemaDump(migrations []Migration) error {
	if m.ctx.MigratorOptions.DumpSchemaFilePath == nil {
		return errors.New("no schema dump file path provided")
	}

	file, err := os.ReadFile(*m.ctx.MigratorOptions.DumpSchemaFilePath)
	if err != nil {
		return fmt.Errorf("unable to read schema dump file: %w", err)
	}

	logger.ShowSQLEvents = false

	tx, err := m.db.BeginTx(m.ctx.Context, nil)
	if err != nil {
		return fmt.Errorf("unable to start transaction: %w", err)
	}

	defer tx.Rollback()

	tx.ExecContext(m.ctx.Context, "SET search_path TO public")
	_, err = tx.ExecContext(m.ctx.Context, string(file))
	if err != nil {
		return fmt.Errorf("unable to apply schema dump: %w", err)
	}

	tx.Commit()

	schema := m.NewSchema()

	versions := make([]string, 0, len(migrations))
	for _, migration := range migrations {
		versions = append(versions, fmt.Sprint(migration.Date().UTC().Format(utils.FormatTime)))
	}

	logger.ShowSQLEvents = false

	schema.AddVersions(versions)

	return nil
}
