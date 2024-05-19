package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"time"
)

type Migration20240518071938CustomSeed struct{}

func (m Migration20240518071938CustomSeed) Change(s *pg.Schema) {
	s.Reversible(schema.Directions{
		Up: func() {
			s.Exec("INSERT INTO migrations_with_change.users (name, email) VALUES ('alexis', 'alexs')")
		},
		Down: func() {
			s.Exec("DELETE FROM migrations_with_change.users WHERE id = '1'")
		},
	})
}

func (m Migration20240518071938CustomSeed) Name() string {
	return "custom_seed"
}

func (m Migration20240518071938CustomSeed) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-05-18T09:19:38+02:00")
	return t
}
