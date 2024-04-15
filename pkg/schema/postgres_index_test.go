package schema

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostgres_AddIndex(t *testing.T) {
	t.Parallel()

	schema := "tst_pg_add_index"

	base := `
CREATE TABLE IF NOT EXISTS {schema}.articles
(
    id   serial PRIMARY KEY,
    name text
);`

	t.Run("simple index", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 0)

		p.AddIndexConstraint(Table("articles", schema), []string{"name"}, IndexOptions{})

		assertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, Table("articles", schema), "idx_articles_name")
	})

	t.Run("with unique index", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 1)

		p.AddIndexConstraint(Table("articles", schema), []string{"name"}, IndexOptions{Unique: true})

		assertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, Table("articles", schema), "idx_articles_name")
	})

	t.Run("with custom name", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 2)

		p.AddIndexConstraint(Table("articles", schema), []string{"name"}, IndexOptions{
			IndexNameBuilder: func(table TableName, columns []string) string {
				return "lalalalala"
			},
		})

		assertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, Table("articles", schema), "lalalalala")
	})

	t.Run("with a method", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 3)

		p.AddIndexConstraint(Table("articles", schema), []string{"name"}, IndexOptions{
			Method: "btree",
		})

		assertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, Table("articles", schema), "idx_articles_name")
	})

	t.Run("with concurrently", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 4)

		p.AddIndexConstraint(Table("articles", schema), []string{"name"}, IndexOptions{
			Concurrent: true,
		})

		assertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, Table("articles", schema), "idx_articles_name")
	})

	t.Run("with order", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 5)

		p.AddIndexConstraint(Table("articles", schema), []string{"name"}, IndexOptions{
			Order: "DESC",
		})

		assertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, Table("articles", schema), "idx_articles_name")
	})

	t.Run("with multiple Columns", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 6)

		p.AddIndexConstraint(Table("articles", schema), []string{"name", "id"}, IndexOptions{})

		assertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, Table("articles", schema), "idx_articles_name_id")
	})

	t.Run("with order per column", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 7)

		p.AddIndexConstraint(Table("articles", schema), []string{"name", "id"}, IndexOptions{
			OrderPerColumn: map[string]string{
				"name": "DESC",
				"id":   "ASC NULLS LAST",
			},
		})

		assertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, Table("articles", schema), "idx_articles_name_id")
	})

	t.Run("with predicate", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 8)

		p.AddIndexConstraint(Table("articles", schema), []string{"name"}, IndexOptions{
			Predicate: "name IS NOT NULL",
		})

		assertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, Table("articles", schema), "idx_articles_name")
	})

	t.Run("with / without IfNotExists", func(t *testing.T) {
		t.Parallel()
		p, _, schema := baseTest(t, base, schema, 9)

		p.AddIndexConstraint(Table("articles", schema), []string{"name"}, IndexOptions{})

		require.Panics(t, func() {
			p.AddIndexConstraint(Table("articles", schema), []string{"name"}, IndexOptions{})
		})

		t.Run("ensure no panic if param IfNotExists is true", func(t *testing.T) {
			p.AddIndexConstraint(Table("articles", schema), []string{"name"}, IndexOptions{IfNotExists: true})
		})
	})
}
