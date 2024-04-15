package schema

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostgres_AddColumn(t *testing.T) {
	t.Parallel()

	schema := "tst_pg_add_column"

	base := `
CREATE TABLE IF NOT EXISTS {schema}.articles()
    `

	t.Run("simple column", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 0)

		p.AddColumn(Table("articles", schema), "name", "text", ColumnOptions{})

		assertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "text"},
		}, dumpColumns(t, p, Table("articles", schema)))
	})

	t.Run("with default value", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 1)

		p.AddColumn(Table("articles", schema), "name", "text", ColumnOptions{
			Default: "default_name",
		})

		assertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "text", ColumnDefault: Ptr("'default_name'::text")},
		}, dumpColumns(t, p, Table("articles", schema)))
	})

	t.Run("varchar limit", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 2)

		p.AddColumn(Table("articles", schema), "name", "varchar", ColumnOptions{
			Limit: 255,
		})

		assertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "character varying"},
		}, dumpColumns(t, p, Table("articles", schema)))

	})

	t.Run("with type primary key", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 3)

		p.AddColumn(Table("articles", schema), "id", ColumnTypePrimaryKey, ColumnOptions{})

		assertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "id", DataType: "integer", PrimaryKey: true, ColumnDefault: Ptr("nextval")},
		}, dumpColumns(t, p, Table("articles", schema)))
	})

	t.Run("with type serial", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 4)

		p.AddColumn(Table("articles", schema), "a", ColumnTypeSerial, ColumnOptions{})
		p.AddColumn(Table("articles", schema), "b", ColumnTypeSerial, ColumnOptions{
			Limit: 10,
		})
		p.AddColumn(Table("articles", schema), "c", ColumnTypeSerial, ColumnOptions{
			Limit: 2,
		})

		assertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "a", DataType: "integer", ColumnDefault: Ptr("nextval")},
			{ColumnName: "b", DataType: "bigint", ColumnDefault: Ptr("nextval")},
			{ColumnName: "c", DataType: "smallint", ColumnDefault: Ptr("nextval")},
		}, dumpColumns(t, p, Table("articles", schema)))
	})

	t.Run("with precision and or scale", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 5)

		p.AddColumn(Table("articles", schema), "a", ColumnTypeDecimal, ColumnOptions{
			Precision: 10,
			Scale:     2,
		})
		p.AddColumn(Table("articles", schema), "b", ColumnTypeNumeric, ColumnOptions{
			Precision: 10,
		})
		p.AddColumn(Table("articles", schema), "c", ColumnTypeDatetime, ColumnOptions{})
		p.AddColumn(Table("articles", schema), "d", ColumnTypeDatetime, ColumnOptions{
			Precision: 8,
		})

		require.PanicsWithError(t, "scale cannot be set without setting the precision", func() {
			p.AddColumn(Table("articles", schema), "e", ColumnTypeDecimal, ColumnOptions{
				Scale: 2,
			})
		})

		assertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "a", DataType: "numeric"},
			{ColumnName: "b", DataType: "numeric"},
			{ColumnName: "c", DataType: "timestamp without time zone"},
			{ColumnName: "d", DataType: "timestamp without time zone"},
		}, dumpColumns(t, p, Table("articles", schema)))
	})

	t.Run("custom column type", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 7)

		p.AddColumn(Table("articles", schema), "id", "serial", ColumnOptions{})
		p.AddColumn(Table("articles", schema), "id_plus_1", "numeric GENERATED ALWAYS AS (id + 1) STORED",
			ColumnOptions{})

		assertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "id", DataType: "integer", ColumnDefault: Ptr("nextval")},
			{ColumnName: "id_plus_1", DataType: "numeric"},
		}, dumpColumns(t, p, Table("articles", schema)))

	})

	t.Run("with array", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 8)

		p.AddColumn(Table("articles", schema), "name", "text", ColumnOptions{
			Array: true,
		})

		p.AddColumn(Table("articles", schema), "tetarraydec", ColumnTypeDecimal, ColumnOptions{
			Array:     true,
			Precision: 10,
			Scale:     2,
		})

		assertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "ARRAY"},
			{ColumnName: "tetarraydec", DataType: "ARRAY"},
		}, dumpColumns(t, p, Table("articles", schema)))
	})

	t.Run("with not null", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 9)

		p.AddColumn(Table("articles", schema), "name", "text", ColumnOptions{
			NotNull: true,
		})

		assertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "text"},
		}, dumpColumns(t, p, Table("articles", schema)))
	})

	t.Run("with if not exists", func(t *testing.T) {
		t.Parallel()
		p, _, schema := baseTest(t, base, schema, 10)

		p.AddColumn(Table("articles", schema), "name", "text", ColumnOptions{})

		require.Panics(t, func() {
			p.AddColumn(Table("articles", schema), "name", "text", ColumnOptions{})
		})

		p.AddColumn(Table("articles", schema), "name", "text", ColumnOptions{IfNotExists: true})
	})

	t.Run("with comment", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 11)

		p.AddColumn(Table("articles", schema), "name", "text", ColumnOptions{
			Comment: "this is a comment",
		})

		assertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "text"},
		}, dumpColumns(t, p, Table("articles", schema)))
	})

}

func TestPostgres_AddColumnComment(t *testing.T) {
	t.Parallel()

	schema := "tst_pg_add_column_comment"

	base := `create table {schema}.articles(id integer, name text);`

	t.Run("simple comment", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 0)

		p.AddColumnComment(Table("articles", schema), "id", Ptr("this is a comment"), ColumnCommentOptions{})

		assertSnapshotDiff(t, r.String())
	})

	t.Run("null comment", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 1)

		p.AddColumnComment(Table("articles", schema), "id", nil, ColumnCommentOptions{})

		assertSnapshotDiff(t, r.String())
	})
}
