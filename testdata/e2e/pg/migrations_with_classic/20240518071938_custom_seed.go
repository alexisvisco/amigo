package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"time"
)

type Migration20240518071938CustomSeed struct{}

func (m Migration20240518071938CustomSeed) Up(s *pg.Schema) {
	s.Exec("INSERT INTO migrations_with_classic.users (name, email) VALUES ('alexis', 'alexs')")
}

func (m Migration20240518071938CustomSeed) Down(s *pg.Schema) {
	s.Exec("DELETE FROM migrations_with_classic.users WHERE name = 'alexis'")
}

func (m Migration20240518071938CustomSeed) Name() string {
	return "custom_seed"
}

func (m Migration20240518071938CustomSeed) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-05-18T09:19:38+02:00")
	return t
}
