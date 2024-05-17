package pg

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils/testutils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostgres_CreateTable(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_create_table"

	t.Run("create basic table", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 0)

		p.CreateTable(schema.Table("articles", sc), func(t *PostgresTableDef) {
			t.Serial("id")
			t.String("title", schema.ColumnOptions{
				Constraints: []schema.ConstraintOption{
					schema.CheckConstraintOptions{
						ConstraintName: "title_not_empty",
						Expression:     "title <> ''",
					},
				},
			})

			t.Text("content")
			t.Integer("views")
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
		assertTableExist(t, p, schema.Table("articles", sc))
	})

	t.Run("with custom primary key name", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 1)

		p.CreateTable(schema.Table("articles", sc), func(t *PostgresTableDef) {
			t.Serial("custom_id", schema.ColumnOptions{
				Limit: 8,
			})
		}, schema.TableOptions{
			PrimaryKeys: []string{"custom_id"},
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
		assertTableExist(t, p, schema.Table("articles", sc))
	})

	t.Run("composite primary key", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 2)

		p.CreateTable(schema.Table("articles", sc), func(t *PostgresTableDef) {
			t.Serial("id")
			t.String("author_id")
			t.Text("content")
			t.Integer("views")
		}, schema.TableOptions{
			PrimaryKeys: []string{"id", "author_id"},
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
		assertTableExist(t, p, schema.Table("articles", sc))
	})

	t.Run("foreign keys", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 3)

		p.CreateTable(schema.Table("articles", sc), func(t *PostgresTableDef) {
			t.Serial("id")
			t.String("author_id")
			t.Text("content")
			t.Integer("views")
		})

		p.CreateTable(schema.Table("authors", sc), func(t *PostgresTableDef) {
			t.Serial("id")
			t.String("name")
			t.Integer("article_id")
			t.ForeignKey(schema.Table("articles", sc))
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
		assertTableExist(t, p, schema.Table("articles", sc))
		assertTableExist(t, p, schema.Table("authors", sc))
	})

	t.Run("indexes", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 4)

		p.CreateTable(schema.Table("articles", sc), func(t *PostgresTableDef) {
			t.Serial("id")
			t.String("title")
			t.Text("content")
			t.Integer("views")
			t.Timestamps()

			t.Index([]string{"title"})
			t.Index([]string{"content", "views"})
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
		assertTableExist(t, p, schema.Table("articles", sc))
	})

	t.Run("without primary key", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 5)

		p.CreateTable(schema.Table("articles", sc), func(t *PostgresTableDef) {
			t.Serial("id")
			t.String("title")
		}, schema.TableOptions{
			WithoutPrimaryKey: true,
		})

		// no need to specify WithoutPrimaryKey: true because there is no id column
		p.CreateTable(schema.Table("articles_without_id", sc), func(t *PostgresTableDef) {
			t.String("title")
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords(), true)
		assertTableExist(t, p, schema.Table("articles", sc))
		assertTableExist(t, p, schema.Table("articles_without_id", sc))
	})

	t.Run("could not find id", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, "select 1;", sc, 6)

		require.PanicsWithError(t, "primary key column ref is not defined", func() {
			p.CreateTable(schema.Table("articles", sc), func(t *PostgresTableDef) {
				t.String("title")
			}, schema.TableOptions{PrimaryKeys: []string{"ref"}})
		})
	})
}

func TestPostgres_DropTable(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_drop_table"

	t.Run("drop table", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 0)

		p.CreateTable(schema.Table("articles", sc), func(t *PostgresTableDef) {
			t.Serial("id")
		})

		p.DropTable(schema.Table("articles", sc))

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
		assertTableNotExist(t, p, schema.Table("articles", sc))
	})

	t.Run("drop table with if exists", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 1)

		require.Panics(t, func() {
			p.DropTable(schema.Table("articles", sc))
		})
		p.DropTable(schema.Table("articles", sc), schema.DropTableOptions{IfExists: true})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
		assertTableNotExist(t, p, schema.Table("articles", sc))
	})
}
