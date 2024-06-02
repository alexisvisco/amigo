package pg

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils/testutils"
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

		p.AddIndex(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name")
	})

	t.Run("with unique index", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 1)

		p.AddIndex(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{Unique: true})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name")
	})

	t.Run("with custom name", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 2)

		p.AddIndex(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{
			IndexNameBuilder: func(table schema.TableName, columns []string) string {
				return "lalalalala"
			},
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, schema.Table("articles", sc), "lalalalala")
	})

	t.Run("with a method", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 3)

		p.AddIndex(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{
			Method: "btree",
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name")
	})

	t.Run("with concurrently", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 4)

		p.AddIndex(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{
			Concurrent: true,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name")
	})

	t.Run("with order", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 5)

		p.AddIndex(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{
			Order: "DESC",
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name")
	})

	t.Run("with multiple Columns", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 6)

		p.AddIndex(schema.Table("articles", sc), []string{"name", "id"}, schema.IndexOptions{})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name_id")
	})

	t.Run("with order per column", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 7)

		p.AddIndex(schema.Table("articles", sc), []string{"name", "id"}, schema.IndexOptions{
			OrderPerColumn: map[string]string{
				"name": "DESC",
				"id":   "ASC NULLS LAST",
			},
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name_id")
	})

	t.Run("with predicate", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 8)

		p.AddIndex(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{
			Predicate: "name IS NOT NULL",
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, schema.Table("articles", sc), "idx_articles_name")
	})

	t.Run("with / without IfNotExists", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, base, sc, 9)

		p.AddIndex(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{})

		require.Panics(t, func() {
			p.AddIndex(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{})
		})

		t.Run("ensure no panic if param IfNotExists is true", func(t *testing.T) {
			p.AddIndex(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{IfNotExists: true})
		})
	})
}

func TestPostgres_DropIndex(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_drop_index"

	base := `create table {schema}.articles (id serial primary key, name text); create index idx_articles_name on {schema}.articles (name);`

	t.Run("simple index", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 0)

		p.DropIndex(schema.Table("articles", sc), []string{"name"})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("with custom name", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 1)

		p.AddIndex(schema.Table("articles", sc), []string{"name"}, schema.IndexOptions{
			IndexName: "idx_articles_name_custom",
		})

		p.DropIndex(schema.Table("articles", sc), []string{"name"}, schema.DropIndexOptions{
			IndexName: "idx_articles_name_custom",
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("if exists", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 2)

		p.DropIndex(schema.Table("articles", sc), []string{"name"}, schema.DropIndexOptions{
			IndexName: "idx_articles_name_custom",
			IfExists:  true,
		})

		require.Panics(t, func() {
			p.DropIndex(schema.Table("articles", sc), []string{"name"}, schema.DropIndexOptions{
				IndexName: "idx_articles_name_custom",
			})
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})
}
