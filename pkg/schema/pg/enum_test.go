package pg

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/testutils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostgres_AddEnum(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_add_enum"

	t.Run("add enum", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 0)

		p.CreateEnum("status", []string{"active", "inactive"}, schema.CreateEnumOptions{
			Schema: sc,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("add enum with no values", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 1)

		p.CreateEnum("status", []string{}, schema.CreateEnumOptions{
			Schema: sc,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})
}

func TestPostgres_AddEnumValue(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_add_enum_value"

	t.Run("add enum value", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 0)

		p.CreateEnum("status", []string{"active", "inactive"}, schema.CreateEnumOptions{
			Schema: sc,
		})

		p.AddEnumValue("status", "pending", schema.AddEnumValueOptions{
			Schema: sc,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("add enum value after/before a value", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 1)

		p.CreateEnum("status", []string{"active", "inactive"}, schema.CreateEnumOptions{
			Schema: sc,
		})

		p.AddEnumValue("status", "pending", schema.AddEnumValueOptions{
			Schema:      sc,
			BeforeValue: "active",
		})

		p.AddEnumValue("status", "rejected", schema.AddEnumValueOptions{
			Schema:     sc,
			AfterValue: "inactive",
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})
}

func TestPostgres_ListEnumValues(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_list_enum_values"

	t.Run("list enum values", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, "select 1;", sc, 0)

		p.CreateEnum("status", []string{"active", "inactive"}, schema.CreateEnumOptions{
			Schema: sc,
		})

		values := p.ListEnumValues("status", utils.Ptr(sc))

		require.Equal(t, []string{"active", "inactive"}, values)
	})
}

func TestPostgres_DropEnum(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_drop_enum"

	t.Run("drop enum", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 0)

		p.CreateEnum("status", []string{"active", "inactive"}, schema.CreateEnumOptions{
			Schema: sc,
		})

		p.DropEnum("status", schema.DropEnumOptions{
			Schema: sc,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("drop enum with if exists", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 1)

		require.Panics(t, func() {
			p.DropEnum("status", schema.DropEnumOptions{
				Schema: sc,
			})
		})

		p.DropEnum("status", schema.DropEnumOptions{
			Schema:   sc,
			IfExists: true,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})

	t.Run("require panics if enum is used", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, "select 1;", sc, 2)

		p.CreateEnum("status", []string{"active", "inactive"}, schema.CreateEnumOptions{
			Schema: sc,
		})

		p.CreateTable(schema.Table("articles", sc), func(t *PostgresTableDef) {
			t.Column("status", "tst_pg_drop_enum_2.status")
		})

		require.Panics(t, func() {
			p.DropEnum("status", schema.DropEnumOptions{
				Schema: sc,
			})
		})
	})
}

func TestPostgres_RenameEnum(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_rename_enum"

	t.Run("rename enum", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 0)

		p.CreateEnum("status", []string{"active", "inactive"}, schema.CreateEnumOptions{
			Schema: sc,
		})

		p.RenameEnum("status", "status2", schema.RenameEnumOptions{
			Schema: sc,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords())
	})
}

func TestPostgres_RenameValue(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_rename_enum_value"

	t.Run("rename enum value", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, "select 1;", sc, 0)

		p.CreateEnum("status", []string{"active", "inactive"}, schema.CreateEnumOptions{
			Schema: sc,
		})

		p.CreateTable(schema.Table("articles", sc), func(t *PostgresTableDef) {
			t.Column("status", "tst_pg_rename_enum_value_0.status")
		})

		// insert some data
		_, err := p.TX.ExecContext(p.Context.Context,
			"INSERT INTO tst_pg_rename_enum_value_0.articles (status) VALUES ('active');")
		require.NoError(t, err)

		p.RenameEnumValue("status", "active", "pending", schema.RenameEnumValueOptions{
			Schema: sc,
		})

		testutils.AssertSnapshotDiff(t, r.FormatRecords(), true)
	})
}

func TestPostgres_FindEnumUsage(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_find_enum_usage"

	t.Run("find enum usage", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, "select 1;", sc, 0)

		p.CreateEnum("status", []string{"active", "inactive"}, schema.CreateEnumOptions{
			Schema: sc,
		})

		p.CreateTable(schema.Table("articles", sc), func(t *PostgresTableDef) {
			t.Column("status", "tst_pg_find_enum_usage_0.status")
		})

		p.CreateTable(schema.Table("users", sc), func(t *PostgresTableDef) {
			t.Column("status1", "tst_pg_find_enum_usage_0.status")
		})

		tables := p.FindEnumUsage("status", utils.Ptr(sc))

		require.Len(t, tables, 2)
		require.Equal(t, schema.Table("articles", sc), tables[0].Table)
		require.Equal(t, "status", tables[0].Column)

		require.Equal(t, schema.Table("users", sc), tables[1].Table)
		require.Equal(t, "status1", tables[1].Column)
	})
}
