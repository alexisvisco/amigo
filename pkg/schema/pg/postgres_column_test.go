package pg

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/testutils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostgres_AddColumn(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_add_column"

	base := `
CREATE TABLE IF NOT EXISTS {schema}.articles()
    `

	t.Run("simple column", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 0)

		p.AddColumn(schema.Table("articles", sc), "name", "text", schema.ColumnOptions{})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "text"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("with default value", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 1)

		p.AddColumn(schema.Table("articles", sc), "name", "text", schema.ColumnOptions{
			Default: "'default_name'",
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "text", ColumnDefault: utils.Ptr("'default_name'::text")},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("varchar limit", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 2)

		p.AddColumn(schema.Table("articles", sc), "name", "varchar", schema.ColumnOptions{
			Limit: 255,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "character varying"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))

	})

	t.Run("with type primary key", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 3)

		p.AddColumn(schema.Table("articles", sc), "id", schema.ColumnTypePrimaryKey, schema.ColumnOptions{})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		require.Equal(t, []columnInfo{
			{ColumnName: "id", DataType: "integer", PrimaryKey: true, ColumnDefault: utils.Ptr("nextval")},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("with type serial", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 4)

		p.AddColumn(schema.Table("articles", sc), "a", schema.ColumnTypeSerial, schema.ColumnOptions{})
		p.AddColumn(schema.Table("articles", sc), "b", schema.ColumnTypeSerial, schema.ColumnOptions{
			Limit: 10,
		})
		p.AddColumn(schema.Table("articles", sc), "c", schema.ColumnTypeSerial, schema.ColumnOptions{
			Limit: 2,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		require.Equal(t, []columnInfo{
			{ColumnName: "a", DataType: "integer", ColumnDefault: utils.Ptr("nextval")},
			{ColumnName: "b", DataType: "bigint", ColumnDefault: utils.Ptr("nextval")},
			{ColumnName: "c", DataType: "smallint", ColumnDefault: utils.Ptr("nextval")},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("with precision and or scale", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 5)

		p.AddColumn(schema.Table("articles", sc), "a", schema.ColumnTypeDecimal, schema.ColumnOptions{
			Precision: 10,
			Scale:     2,
		})
		p.AddColumn(schema.Table("articles", sc), "b", schema.ColumnTypeNumeric, schema.ColumnOptions{
			Precision: 10,
		})
		p.AddColumn(schema.Table("articles", sc), "c", schema.ColumnTypeDatetime, schema.ColumnOptions{})
		p.AddColumn(schema.Table("articles", sc), "d", schema.ColumnTypeDatetime, schema.ColumnOptions{
			Precision: 8,
		})

		require.PanicsWithError(t, "scale cannot be set without setting the precision", func() {
			p.AddColumn(schema.Table("articles", sc), "e", schema.ColumnTypeDecimal, schema.ColumnOptions{
				Scale: 2,
			})
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		require.Equal(t, []columnInfo{
			{ColumnName: "a", DataType: "numeric"},
			{ColumnName: "b", DataType: "numeric"},
			{ColumnName: "c", DataType: "timestamp without time zone"},
			{ColumnName: "d", DataType: "timestamp without time zone"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("custom column type", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 7)

		p.AddColumn(schema.Table("articles", sc), "id", "serial", schema.ColumnOptions{})
		p.AddColumn(schema.Table("articles", sc), "id_plus_1", "numeric GENERATED ALWAYS AS (id + 1) STORED",
			schema.ColumnOptions{})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		require.Equal(t, []columnInfo{
			{ColumnName: "id", DataType: "integer", ColumnDefault: utils.Ptr("nextval")},
			{ColumnName: "id_plus_1", DataType: "numeric"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))

	})

	t.Run("with array", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 8)

		p.AddColumn(schema.Table("articles", sc), "name", "text", schema.ColumnOptions{
			Array: true,
		})

		p.AddColumn(schema.Table("articles", sc), "tetarraydec", schema.ColumnTypeDecimal, schema.ColumnOptions{
			Array:     true,
			Precision: 10,
			Scale:     2,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "ARRAY"},
			{ColumnName: "tetarraydec", DataType: "ARRAY"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("with not null", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 9)

		p.AddColumn(schema.Table("articles", sc), "name", "text", schema.ColumnOptions{
			NotNull: true,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "text"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("with if not exists", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, base, sc, 10)

		p.AddColumn(schema.Table("articles", sc), "name", "text", schema.ColumnOptions{})

		require.Panics(t, func() {
			p.AddColumn(schema.Table("articles", sc), "name", "text", schema.ColumnOptions{})
		})

		p.AddColumn(schema.Table("articles", sc), "name", "text", schema.ColumnOptions{IfNotExists: true})
	})

	t.Run("with comment", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 11)

		p.AddColumn(schema.Table("articles", sc), "name", "text", schema.ColumnOptions{
			Comment: "this is a comment",
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "text"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("with timestamps", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 12)

		p.AddTimestamps(schema.Table("articles", sc))

		testutils.AssertSnapshotDiff(t, r.FormatRecords(), true)
	})

}

func TestPostgres_DropColumn(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_drop_column"

	base := `create table {schema}.articles(id integer, name text);`

	t.Run("simple drop", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 0)

		p.DropColumn(schema.Table("articles", sc), "id")

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "text"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("drop if exists", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 1)

		p.DropColumn(schema.Table("articles", sc), "id", schema.DropColumnOptions{IfExists: true})
		p.DropColumn(schema.Table("articles", sc), "id", schema.DropColumnOptions{IfExists: true})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())

		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "text"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})
}

func TestPostgres_AddColumnComment(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_add_column_comment"

	base := `create table {schema}.articles(id integer, name text);`

	t.Run("simple comment", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 0)

		p.AddColumnComment(schema.Table("articles", sc), "id", utils.Ptr("this is a comment"),
			schema.ColumnCommentOptions{})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("null comment", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 1)

		p.AddColumnComment(schema.Table("articles", sc), "id", nil, schema.ColumnCommentOptions{})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})
}

func TestPostgres_RenameColumn(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_rename_column"

	base := `create table {schema}.articles(id integer, name text);`

	t.Run("simple rename", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 0)

		p.RenameColumn(schema.Table("articles", sc), "id", "new_id")

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
		require.Equal(t, []columnInfo{
			{ColumnName: "name", DataType: "text"},
			{ColumnName: "new_id", DataType: "integer"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})
}

func TestPostgres_ChangeColumn(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_change_column"

	base := `create table {schema}.articles(id integer, name text);`

	t.Run("simple change", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 0)

		p.ChangeColumnType(schema.Table("articles", sc), "name", "varchar", schema.ChangeColumnTypeOptions{
			Limit: 255,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
		require.Equal(t, []columnInfo{
			{ColumnName: "id", DataType: "integer"},
			{ColumnName: "name", DataType: "character varying"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("change column type with using", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 1)

		p.ChangeColumnType(schema.Table("articles", sc), "name", "integer", schema.ChangeColumnTypeOptions{
			Using: "length(name)", // rly dumb example but it's just for the test
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
		require.Equal(t, []columnInfo{
			{ColumnName: "id", DataType: "integer"},
			{ColumnName: "name", DataType: "integer"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})
}
