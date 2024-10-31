package migrations

import (
	"time"

	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/pg"
)

type Migration20240517080505SchemaVersion struct{}

func (m Migration20240517080505SchemaVersion) Up(s *pg.Schema) {
	s.CreateTable("migrations_with_classic.mig_schema_versions", func(s *pg.PostgresTableDef) {
		s.String("version", schema.ColumnOptions{PrimaryKey: true})
	}, schema.TableOptions{IfNotExists: true})
}

func (m Migration20240517080505SchemaVersion) Down(s *pg.Schema) {
	s.DropTable("migrations_with_classic.mig_schema_versions")
}

func (m Migration20240517080505SchemaVersion) Name() string {
	return "schema_version"
}

func (m Migration20240517080505SchemaVersion) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-05-17T10:05:05+02:00")
	return t
}
