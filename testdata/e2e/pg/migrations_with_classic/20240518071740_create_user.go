package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"time"
)

type Migration20240518071740CreateUser struct{}

func (m Migration20240518071740CreateUser) Up(s *pg.Schema) {
	s.CreateTable("migrations_with_classic.users", func(def *pg.PostgresTableDef) {
		def.Serial("id")
		def.String("name")
		def.String("email")
		def.Timestamps()
		def.Index([]string{"name"})
	})
}

func (m Migration20240518071740CreateUser) Down(s *pg.Schema) {
	s.DropTable("migrations_with_classic.users")
}

func (m Migration20240518071740CreateUser) Name() string {
	return "create_user"
}

func (m Migration20240518071740CreateUser) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-05-18T09:17:40+02:00")
	return t
}
