package sqlite

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils/testutils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSQLite_AddIndex(t *testing.T) {
	t.Parallel()

	base := `
CREATE TABLE IF NOT EXISTS articles
(
    id   serial PRIMARY KEY,
    name text
);`

	t.Run("simple index", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.AddIndex("articles", []string{"name"}, schema.IndexOptions{})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, "articles", "idx_articles_name")
	})

	t.Run("with unique index", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.AddIndex("articles", []string{"name"}, schema.IndexOptions{Unique: true})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, "articles", "idx_articles_name")
	})

	t.Run("with custom name", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.AddIndex("articles", []string{"name"}, schema.IndexOptions{
			IndexNameBuilder: func(table schema.TableName, columns []string) string {
				return "lalalalala"
			},
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, "articles", "lalalalala")
	})

	t.Run("with order", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.AddIndex("articles", []string{"name"}, schema.IndexOptions{
			Order: "DESC",
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, "articles", "idx_articles_name")
	})

	t.Run("with multiple Columns", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.AddIndex("articles", []string{"name", "id"}, schema.IndexOptions{})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, "articles", "idx_articles_name_id")
	})

	t.Run("with order per column", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.AddIndex("articles", []string{"name", "id"}, schema.IndexOptions{
			OrderPerColumn: map[string]string{
				"name": "DESC",
				"id":   "ASC",
			},
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, "articles", "idx_articles_name_id")
	})

	t.Run("with predicate", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.AddIndex("articles", []string{"name"}, schema.IndexOptions{
			Predicate: "name IS NOT NULL",
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		asserIndexExist(t, p, "articles", "idx_articles_name")
	})

	t.Run("with / without IfNotExists", func(t *testing.T) {
		t.Parallel()
		p, _ := baseTest(t, base)

		p.AddIndex("articles", []string{"name"}, schema.IndexOptions{})

		require.Panics(t, func() {
			p.AddIndex("articles", []string{"name"}, schema.IndexOptions{})
		})

		t.Run("ensure no panic if param IfNotExists is true", func(t *testing.T) {
			p.AddIndex("articles", []string{"name"}, schema.IndexOptions{IfNotExists: true})
		})
	})
}

func TestSQLite_DropIndex(t *testing.T) {
	t.Parallel()

	testutils.EnableSnapshotForAll()

	base := `create table articles (id serial primary key, name text); create index idx_articles_name on articles (name);`

	t.Run("simple index", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.DropIndex("articles", []string{"name"})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("with custom name", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.AddIndex("articles", []string{"name"}, schema.IndexOptions{
			IndexName: "idx_articles_name_custom",
		})

		p.DropIndex("articles", []string{"name"}, schema.DropIndexOptions{
			IndexName: "idx_articles_name_custom",
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("if exists", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.DropIndex("articles", []string{"name"}, schema.DropIndexOptions{
			IndexName: "idx_articles_name_custom",
			IfExists:  true,
		})

		require.Panics(t, func() {
			p.DropIndex("articles", []string{"name"}, schema.DropIndexOptions{
				IndexName: "idx_articles_name_custom",
			})
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})
}
