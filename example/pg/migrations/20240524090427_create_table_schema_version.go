package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"time"
)

type Migration20240524090427CreateTableSchemaVersion struct{}

func (m Migration20240524090427CreateTableSchemaVersion) Up(s *pg.Schema) {
	s.CreateTable("public.mig_schema_versions", func(s *pg.PostgresTableDef) {
		s.String("id")
	}, schema.TableOptions{IfNotExists: true})
}

func (m Migration20240524090427CreateTableSchemaVersion) Down(s *pg.Schema) {}

func (m Migration20240524090427CreateTableSchemaVersion) Name() string {
	return "create_table_schema_version"
}

func (m Migration20240524090427CreateTableSchemaVersion) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-05-24T11:04:27+02:00")
	return t
}
