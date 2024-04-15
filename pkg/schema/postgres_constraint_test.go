package schema

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostgres_AddCheckConstraint(t *testing.T) {
	t.Parallel()
	schema := "tst_pg_add_check_constraint"

	sql := `CREATE TABLE IF NOT EXISTS {schema}.test_table (
			id serial PRIMARY KEY,
			name text,
           not_valid_but_exists text
		);`

	t.Run("with Table prefix", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, sql, schema, 1)

		p.AddCheckConstraint(Table("test_table", schema),
			"constraint_1",
			"name <> ''",
			CheckConstraintOptions{})

		assertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, Table("test_table", schema), "ck_test_table_constraint_1")
	})

	t.Run("with custom name", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, sql, schema, 2)

		p.AddCheckConstraint(Table("test_table", schema),
			"constraint_2",
			"name <> ''",
			CheckConstraintOptions{
				ConstraintNameBuilder: func(tableName TableName, constraintName string) string {
					return "lalalalala"
				},
			})

		assertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, Table("test_table", schema), "lalalalala")
	})

	t.Run("with no validate", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, sql, schema, 3)

		p.AddCheckConstraint(Table("test_table", schema),
			"constraint_3",
			"name <> ''",
			CheckConstraintOptions{Validate: Ptr(false)})

		assertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, Table("test_table", schema), "ck_test_table_constraint_3")
	})

	t.Run("with / without IfNotExists", func(t *testing.T) {
		t.Parallel()
		p, _, schema := baseTest(t, sql, schema, 4)

		p.AddCheckConstraint(Table("test_table", schema),
			"constraint_4",
			"name <> ''",
			CheckConstraintOptions{})

		require.Panics(t, func() {
			p.AddCheckConstraint(Table("test_table", schema),
				"constraint_4",
				"name <> ''",
				CheckConstraintOptions{})
		})

		t.Run("ensure no panic if param IfNotExists is true", func(t *testing.T) {
			p.AddCheckConstraint(Table("test_table", schema),
				"constraint_4",
				"name <> ''",
				CheckConstraintOptions{IfNotExists: true})
		})
	})
}

func TestPostgres_AddForeignKeyConstraint(t *testing.T) {
	t.Parallel()

	schema := "tst_pg_add_foreign_key_constraint"

	base := `
CREATE TABLE IF NOT EXISTS {schema}.articles
(
    id   serial PRIMARY KEY,
    name text,
	author_id integer
);

CREATE TABLE IF NOT EXISTS {schema}.authors
(
    id      serial PRIMARY KEY,
    name    text
);`

	t.Run("with Table prefix", func(t *testing.T) {
		t.Parallel()
		p, _, schema := baseTest(t, base, schema, 0)

		p.AddForeignKeyConstraint(Table("articles", schema), Table("authors", schema), AddForeignKeyOptions{})
		assertConstraintExist(t, p, Table("articles", schema), "fk_articles_authors")
	})

	t.Run("with custom name", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 1)

		p.AddForeignKeyConstraint(Table("articles", schema), Table("authors", schema), AddForeignKeyOptions{
			ForeignKeyNameBuilder: func(fromTable TableName, toTable TableName) string {
				return "lalalalala"
			},
		})

		assertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, Table("articles", schema), "lalalalala")
	})

	t.Run("custom column name", func(t *testing.T) {
		t.Parallel()
		sql := `
CREATE TABLE IF NOT EXISTS {schema}.articles
(
    id   serial PRIMARY KEY,
    name text,
	user_id integer
);

CREATE TABLE IF NOT EXISTS {schema}.authors
(
    id      serial PRIMARY KEY,
    name    text
);`

		p, r, schema := baseTest(t, sql, schema, 2)

		p.AddForeignKeyConstraint(Table("articles", schema), Table("authors", schema), AddForeignKeyOptions{
			Column: "user_id",
		})

		assertSnapshotDiff(t, r.String())
	})

	t.Run("custom primary key", func(t *testing.T) {
		t.Parallel()
		sql := `
CREATE TABLE IF NOT EXISTS {schema}.articles
(
    id   serial PRIMARY KEY,
    name text,
	user_id integer
);

CREATE TABLE IF NOT EXISTS {schema}.authors
(
    ref      serial PRIMARY KEY,
    name    text
);`

		p, r, schema := baseTest(t, sql, schema, 3)

		p.AddForeignKeyConstraint(Table("articles", schema), Table("authors", schema), AddForeignKeyOptions{
			Column:     "user_id",
			PrimaryKey: "ref",
		})

		assertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, Table("articles", schema), "fk_articles_authors")
	})

	t.Run("composite pk", func(t *testing.T) {
		t.Parallel()
		sql := `
CREATE TABLE IF NOT EXISTS {schema}.carts
(
    shop_id integer,
    user_id integer,
    PRIMARY KEY (shop_id, user_id)
);

CREATE TABLE IF NOT EXISTS {schema}.orders
(
    id serial PRIMARY KEY,
    cart_shop_id integer,
    cart_user_id integer
);`

		p, r, schema := baseTest(t, sql, schema, 4)

		p.AddForeignKeyConstraint(Table("orders", schema), Table("carts", schema), AddForeignKeyOptions{
			CompositePrimaryKey: []string{"shop_id", "user_id"},
		})

		assertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, Table("orders", schema), "fk_orders_carts")
	})

	t.Run("with on delete", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 5)

		p.AddForeignKeyConstraint(Table("articles", schema), Table("authors", schema), AddForeignKeyOptions{
			OnDelete: "cascade",
		})

		assertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, Table("articles", schema), "fk_articles_authors")
	})

	t.Run("with on update", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 6)

		p.AddForeignKeyConstraint(Table("articles", schema), Table("authors", schema), AddForeignKeyOptions{
			OnUpdate: "cascade",
		})

		assertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, Table("articles", schema), "fk_articles_authors")
	})

	t.Run("with deferrable", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 7)

		p.AddForeignKeyConstraint(Table("articles", schema), Table("authors", schema), AddForeignKeyOptions{
			Deferrable: "DEFERRABLE INITIALLY DEFERRED",
		})

		assertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, Table("articles", schema), "fk_articles_authors")
	})

	t.Run("with no validate", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 8)

		p.AddForeignKeyConstraint(Table("articles", schema), Table("authors", schema), AddForeignKeyOptions{
			Validate: Ptr(false),
		})

		assertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, Table("articles", schema), "fk_articles_authors")
	})

	t.Run("with / without IfNotExists", func(t *testing.T) {
		t.Parallel()
		p, _, schema := baseTest(t, base, schema, 9)

		p.AddForeignKeyConstraint(Table("articles", schema), Table("authors", schema), AddForeignKeyOptions{})

		require.Panics(t, func() {
			p.AddForeignKeyConstraint(Table("articles", schema), Table("authors", schema), AddForeignKeyOptions{})
		})

		t.Run("ensure no panic if param IfNotExists is true", func(t *testing.T) {
			p.AddForeignKeyConstraint(Table("articles", schema), Table("authors", schema),
				AddForeignKeyOptions{IfNotExists: true})
		})
	})
}

func TestPostgres_AddPrimaryKeyConstraint(t *testing.T) {
	t.Parallel()

	schema := "tst_pg_add_primary_key_constraint"

	base := `create table {schema}.articles(id integer, name text);`

	t.Run("simple primary key", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 0)

		p.AddPrimaryKeyConstraint(Table("articles", schema), []string{"id"}, PrimaryKeyOptions{})

		assertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "id", DataType: "integer", PrimaryKey: true},
			{ColumnName: "name", DataType: "text"},
		}, dumpColumns(t, p, Table("articles", schema)))
	})

	t.Run("composite primary key", func(t *testing.T) {
		t.Parallel()
		p, r, schema := baseTest(t, base, schema, 1)

		p.AddPrimaryKeyConstraint(Table("articles", schema), []string{"id", "name"}, PrimaryKeyOptions{})

		assertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "id", DataType: "integer", PrimaryKey: true},
			{ColumnName: "name", DataType: "text", PrimaryKey: true},
		}, dumpColumns(t, p, Table("articles", schema)))
	})

	t.Run("if not exists", func(t *testing.T) {
		t.Parallel()
		p, _, schema := baseTest(t, base, schema, 2)

		p.AddPrimaryKeyConstraint(Table("articles", schema), []string{"id"}, PrimaryKeyOptions{})
		require.Panics(t, func() {
			p.AddPrimaryKeyConstraint(Table("articles", schema), []string{"id"}, PrimaryKeyOptions{})
		})
		p.AddPrimaryKeyConstraint(Table("articles", schema), []string{"id"}, PrimaryKeyOptions{IfNotExists: true})
	})
}
