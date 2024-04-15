package migrations

import (
	"github.com/alexisvisco/mig/pkg/schema"
	"time"
)

type MigrationNewTable struct{}

func (m MigrationNewTable) Change(t schema.Postgres) {
	t.AddForeignKeyConstraint("users", "articles", schema.AddForeignKeyOptions{})
	t.AddCheckConstraint(schema.Table("users", "myschema"), "constraint_1", "name <> ''",
		schema.CheckConstraintOptions{})

	t.Reversible(schema.Directions{
		Up: func() {
			// Add a thing here
		},
		Down: func() {
			// reverse the thing here
		},
	})

}

func (m MigrationNewTable) Name() string {
	return "new_table"
}

func (m MigrationNewTable) CreatedDate() (time.Time, error) {
	return time.Parse(time.RFC3339, "2021-08-01T00:00:00Z")
}
