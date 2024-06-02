package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/sqlite"
	"time"
)

type Migration20240602081806DropIndex struct{}

func (m Migration20240602081806DropIndex) Change(s *sqlite.Schema) {
	s.DropIndex("mig_schema_versions", []string{"id"}, schema.DropIndexOptions{
		Reversible: &schema.IndexOptions{IfNotExists: true},
	})
}

func (m Migration20240602081806DropIndex) Name() string {
	return "drop_index"
}

func (m Migration20240602081806DropIndex) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-06-02T10:18:06+02:00")
	return t
}
