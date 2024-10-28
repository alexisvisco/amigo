package sqlite

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils/testutils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSQLite_AddColumn(t *testing.T) {
	t.Parallel()

	testutils.EnableSnapshotForAll()

	base := `
CREATE TABLE IF NOT EXISTS articles
(
    name text
);`

	t.Run("simple column", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.AddColumn("articles", "content", schema.ColumnTypeText)

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("with default value", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.AddColumn("articles", "content", schema.ColumnTypeText, schema.ColumnOptions{Default: "'default content'"})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("with varchar limit", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		p.AddColumn("articles", "content", "varchar", schema.ColumnOptions{Limit: 255})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("with primary key", func(t *testing.T) {
		t.Parallel()
		p, r := baseTest(t, base)

		require.Panics(t, func() {
			p.AddColumn("articles", "id", schema.ColumnTypePrimaryKey, schema.ColumnOptions{PrimaryKey: true})
		})
		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})
}
