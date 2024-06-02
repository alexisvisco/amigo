package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/sqlite"
	"time"
)

type Migration20240602081304AddIndex struct{}

func (m Migration20240602081304AddIndex) Change(s *sqlite.Schema) {
	s.AddIndex("mig_schema_versions", []string{"id"}, schema.IndexOptions{
		IfNotExists: true,
	})
}

func (m Migration20240602081304AddIndex) Name() string {
	return "add_index"
}

func (m Migration20240602081304AddIndex) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-06-02T10:13:04+02:00")
	return t
}
