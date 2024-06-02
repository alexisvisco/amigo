package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema/sqlite"
	"time"
)

type Migration20240602080728CreateTableSchemaVersion struct{}

func (m Migration20240602080728CreateTableSchemaVersion) Up(s *sqlite.Schema) {
	s.Exec("CREATE TABLE IF NOT EXISTS mig_schema_versions ( id VARCHAR(255) NOT NULL PRIMARY KEY )")
}

func (m Migration20240602080728CreateTableSchemaVersion) Down(s *sqlite.Schema) {
	// nothing to do to keep the schema version table
}

func (m Migration20240602080728CreateTableSchemaVersion) Name() string {
	return "create_table_schema_version"
}

func (m Migration20240602080728CreateTableSchemaVersion) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-06-02T10:07:28+02:00")
	return t
}
