package migrations

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema/base"
	"time"
)

type Migration20240529125357CreateTableSchemaVersion struct{}

func (m Migration20240529125357CreateTableSchemaVersion) Up(s *base.Schema) {
	query := `CREATE TABLE IF NOT EXISTS mig_schema_versions ( id VARCHAR(255) NOT NULL PRIMARY KEY )`
	_, err := s.TX.ExecContext(s.Context.Context, query)
	if err != nil {
		s.Context.RaiseError(fmt.Errorf("unable to create table schema version: %w", err))
	}
}

func (m Migration20240529125357CreateTableSchemaVersion) Down(s *base.Schema) {
	// nothing to do to keep the schema version table
}

func (m Migration20240529125357CreateTableSchemaVersion) Name() string {
	return "create_table_schema_version"
}

func (m Migration20240529125357CreateTableSchemaVersion) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-05-29T14:53:57+02:00")
	return t
}
