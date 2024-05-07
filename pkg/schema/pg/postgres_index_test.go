package pg

import (
	"github.com/alexisvisco/mig/pkg/schema"
	"github.com/alexisvisco/mig/pkg/utils/testutils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostgres_AddIndex(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_add_index"

	base := `
CREATE TABLE IF NOT EXISTS {schema}.articles
(
    id   serial PRIMARY KEY,
    name text
);`

	t.Run("simple index", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 0)

		p.AddIndexConstraint(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{})

		testutils.AssertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name")
	})

	t.Run("with unique index", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 1)

		p.AddIndexConstraint(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{Unique: true})

		testutils.AssertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name")
	})

	t.Run("with custom name", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 2)

		p.AddIndexConstraint(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{
			IndexNameBuilder: func(table schema.TableName, columns []string) string {
				return "lalalalala"
			},
		})

		testutils.AssertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, schema.Table("articles", sc), "lalalalala")
	})

	t.Run("with a method", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 3)

		p.AddIndexConstraint(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{
			Method: "btree",
		})

		testutils.AssertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name")
	})

	t.Run("with concurrently", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 4)

		p.AddIndexConstraint(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{
			Concurrent: true,
		})

		testutils.AssertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name")
	})

	t.Run("with order", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 5)

		p.AddIndexConstraint(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{
			Order: "DESC",
		})

		testutils.AssertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name")
	})

	t.Run("with multiple Columns", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 6)

		p.AddIndexConstraint(schema.Table("articles", sc), []string{"name", "id"}, schema.IndexOptions{})

		testutils.AssertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name_id")
	})

	t.Run("with order per column", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 7)

		p.AddIndexConstraint(schema.Table("articles", sc), []string{"name", "id"}, schema.IndexOptions{
			OrderPerColumn: map[string]string{
				"name": "DESC",
				"id":   "ASC NULLS LAST",
			},
		})

		testutils.AssertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name_id")
	})

	t.Run("with predicate", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 8)

		p.AddIndexConstraint(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{
			Predicate: "name IS NOT NULL",
		})

		testutils.AssertSnapshotDiff(t, r.String())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name")
	})

	t.Run("with / without IfNotExists", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, base, sc, 9)

		p.AddIndexConstraint(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{})

		require.Panics(t, func() {
			p.AddIndexConstraint(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{})
		})

		t.Run("ensure no panic if param IfNotExists is true", func(t *testing.T) {
			p.AddIndexConstraint(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{IfNotExists: true})
		})
	})
}
