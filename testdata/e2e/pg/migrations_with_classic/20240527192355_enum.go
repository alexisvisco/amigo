package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"time"
)

type Migration20240527192355Enum struct{}

func (m Migration20240527192355Enum) Up(s *pg.Schema) {
	s.CreateEnum("status", []string{"active", "inactive"}, schema.CreateEnumOptions{
		Schema: "migrations_with_classic",
	})
}

func (m Migration20240527192355Enum) Down(s *pg.Schema) {
	s.DropEnum("status", schema.DropEnumOptions{Schema: "migrations_with_classic"})
}

func (m Migration20240527192355Enum) Name() string {
	return "enum"
}

func (m Migration20240527192355Enum) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-05-27T21:23:55+02:00")
	return t
}
