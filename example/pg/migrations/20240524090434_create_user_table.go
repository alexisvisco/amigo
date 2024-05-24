package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"time"
)

type Migration20240524090434CreateUserTable struct{}

func (m Migration20240524090434CreateUserTable) Change(s *pg.Schema) {
	s.CreateTable("users", func(def *pg.PostgresTableDef) {
		def.AddColumn("id", "bigserial")
		def.String("name")
		def.String("email")
		def.Timestamps()
		def.Index([]string{"name"})
	})
}

func (m Migration20240524090434CreateUserTable) Name() string {
	return "create_user_table"
}

func (m Migration20240524090434CreateUserTable) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-05-24T11:04:34+02:00")
	return t
}
