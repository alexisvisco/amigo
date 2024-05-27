package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"time"
)

type Migration20240527192300Enum struct{}

func (m Migration20240527192300Enum) Change(s *pg.Schema) {
	s.CreateEnum("status", []string{"active", "inactive"}, schema.CreateEnumOptions{
		Schema: "migrations_with_change",
	})
}

func (m Migration20240527192300Enum) Name() string {
	return "enum"
}

func (m Migration20240527192300Enum) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-05-27T21:23:00+02:00")
	return t
}
