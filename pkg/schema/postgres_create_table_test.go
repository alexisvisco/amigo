package schema

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostgresSchema_CreateTable(t *testing.T) {
	t.Parallel()

	schema := "tst_pg_create_table"

	t.Run("create basic table", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, "select 1;", schema, 0)

		p.CreateTable(Table("articles", schema), func(t *PostgresTableDef) {
			t.Serial("id")
			t.String("title", ColumnOptions{
				Constraints: []ConstraintOption{
					CheckConstraintOptions{
						ConstraintName: "title_not_empty",
						Expression:     "title <> ''",
					},
				},
			})

			t.Text("content")
			t.Integer("views")
		})

		assertSnapshotDiff(t, r.String(), true)
		assertTableExist(t, p, Table("articles", schema))
	})

	t.Run("with custom primary key name", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, "select 1;", schema, 1)

		p.CreateTable(Table("articles", schema), func(t *PostgresTableDef) {
			t.Serial("custom_id", ColumnOptions{
				Limit: 8,
			})
		}, TableOptions{
			PrimaryKeys: []string{"custom_id"},
		})

		assertSnapshotDiff(t, r.String(), true)
		assertTableExist(t, p, Table("articles", schema))
	})

	t.Run("composite primary key", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, "select 1;", schema, 2)

		p.CreateTable(Table("articles", schema), func(t *PostgresTableDef) {
			t.Serial("id")
			t.String("author_id")
			t.Text("content")
			t.Integer("views")
		}, TableOptions{
			PrimaryKeys: []string{"id", "author_id"},
		})

		assertSnapshotDiff(t, r.String(), true)
		assertTableExist(t, p, Table("articles", schema))
	})

	t.Run("foreign keys", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, "select 1;", schema, 3)

		p.CreateTable(Table("articles", schema), func(t *PostgresTableDef) {
			t.Serial("id")
			t.String("author_id")
			t.Text("content")
			t.Integer("views")
		})

		p.CreateTable(Table("authors", schema), func(t *PostgresTableDef) {
			t.Serial("id")
			t.String("name")
			t.Integer("article_id")
			t.ForeignKey(Table("articles", schema))
		})

		assertSnapshotDiff(t, r.String(), true)
		assertTableExist(t, p, Table("articles", schema))
		assertTableExist(t, p, Table("authors", schema))
	})

	t.Run("indexes", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, "select 1;", schema, 4)

		p.CreateTable(Table("articles", schema), func(t *PostgresTableDef) {
			t.Serial("id")
			t.String("title")
			t.Text("content")
			t.Integer("views")
			t.Timestamps()

			t.Index([]string{"title"})
			t.Index([]string{"content", "views"})
		})

		assertSnapshotDiff(t, r.String(), true)
		assertTableExist(t, p, Table("articles", schema))
	})
}

func assertTableExist(t *testing.T, p *Postgres, table TableName) {
	var exists bool
	err := p.db.QueryRowContext(context.Background(), `SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = $1
		AND table_name = $2
	);`, table.Schema(), table.Name()).Scan(&exists)

	require.NoError(t, err)
	require.True(t, exists)
}
