package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"time"
)

type Migration20240518071842AddIndexUserEmail struct{}

func (m Migration20240518071842AddIndexUserEmail) Up(s *pg.Schema) {
	s.AddIndexConstraint("migrations_with_classic.users", []string{"email"}, schema.IndexOptions{Unique: true})
}

func (m Migration20240518071842AddIndexUserEmail) Down(s *pg.Schema) {
	s.DropIndex("migrations_with_classic.users", []string{"email"})
}

func (m Migration20240518071842AddIndexUserEmail) Name() string {
	return "add_index_user_email"
}

func (m Migration20240518071842AddIndexUserEmail) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-05-18T09:18:42+02:00")
	return t
}
