package pg

import (
	"github.com/alexisvisco/mig/pkg/schema"
	"github.com/alexisvisco/mig/pkg/utils"
	"github.com/alexisvisco/mig/pkg/utils/testutils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostgres_AddCheckConstraint(t *testing.T) {
	t.Parallel()
	sc := "tst_pg_add_check_constraint"

	sql := `CREATE TABLE IF NOT EXISTS {schema}.test_table (
			id serial PRIMARY KEY,
			name text,
           not_valid_but_exists text
		);`

	t.Run("with Table prefix", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, sql, sc, 1)

		p.AddCheckConstraint(schema.Table("test_table", sc),
			"constraint_1",
			"name <> ''",
			schema.CheckConstraintOptions{})

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, schema.Table("test_table", sc), "ck_test_table_constraint_1")
	})

	t.Run("with custom name", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, sql, sc, 2)

		p.AddCheckConstraint(schema.Table("test_table", sc),
			"constraint_2",
			"name <> ''",
			schema.CheckConstraintOptions{
				ConstraintNameBuilder: func(tableName schema.TableName, constraintName string) string {
					return "lalalalala"
				},
			})

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, schema.Table("test_table", sc), "lalalalala")
	})

	t.Run("with no validate", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, sql, sc, 3)

		p.AddCheckConstraint(schema.Table("test_table", sc),
			"constraint_3",
			"name <> ''",
			schema.CheckConstraintOptions{Validate: utils.Ptr(false)})

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, schema.Table("test_table", sc), "ck_test_table_constraint_3")
	})

	t.Run("with / without IfNotExists", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, sql, sc, 4)

		p.AddCheckConstraint(schema.Table("test_table", sc),
			"constraint_4",
			"name <> ''",
			schema.CheckConstraintOptions{})

		require.Panics(t, func() {
			p.AddCheckConstraint(schema.Table("test_table", sc),
				"constraint_4",
				"name <> ''",
				schema.CheckConstraintOptions{})
		})

		t.Run("ensure no panic if param IfNotExists is true", func(t *testing.T) {
			p.AddCheckConstraint(schema.Table("test_table", sc),
				"constraint_4",
				"name <> ''",
				schema.CheckConstraintOptions{IfNotExists: true})
		})
	})
}

func TestPostgres_DropCheckConstraint(t *testing.T) {
	t.Parallel()
	sc := "tst_pg_drop_check_constraint"

	sql := `CREATE TABLE IF NOT EXISTS {schema}.test_table (
			id serial PRIMARY KEY,
			name text,
			CONSTRAINT ck_test_table_constraint_1 CHECK (name <> '')
		);`

	t.Run("with Table prefix", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, sql, sc, 0)

		p.DropCheckConstraint(schema.Table("test_table", sc), "constraint_1")

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintNotExist(t, p, schema.Table("test_table", sc), "table_constraint_1")
	})

	t.Run("with custom name", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, sql, sc, 1)

		p.DropCheckConstraint(schema.Table("test_table", sc), "", schema.DropCheckConstraintOptions{
			ConstraintName: "ck_test_table_constraint_1",
		})

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintNotExist(t, p, schema.Table("test_table", sc), "ck_test_table_constraint_1")
	})

	t.Run("with / without IfExists", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, sql, sc, 2)

		p.DropCheckConstraint(schema.Table("test_table", sc), "constraint_1")
		require.Panics(t, func() {
			p.DropCheckConstraint(schema.Table("test_table", sc), "constraint_1")
		})

		t.Run("ensure no panic if param IfExists is true", func(t *testing.T) {
			p.DropCheckConstraint(schema.Table("test_table", sc), "constraint_1",
				schema.DropCheckConstraintOptions{IfExists: true})
		})
	})
}

func TestPostgres_AddForeignKeyConstraint(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_add_foreign_key_constraint"

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
		p, _, sc := baseTest(t, base, sc, 0)

		p.AddForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
			schema.AddForeignKeyConstraintOptions{})
		assertConstraintExist(t, p, schema.Table("articles", sc), "fk_articles_authors")
	})

	t.Run("with custom name", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 1)

		p.AddForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
			schema.AddForeignKeyConstraintOptions{
				ForeignKeyNameBuilder: func(fromTable schema.TableName, toTable schema.TableName) string {
					return "lalalalala"
				},
			})

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, schema.Table("articles", sc), "lalalalala")
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

		p, r, sc := baseTest(t, sql, sc, 2)

		p.AddForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
			schema.AddForeignKeyConstraintOptions{
				Column: "user_id",
			})

		testutils.AssertSnapshotDiff(t, r.String())
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

		p, r, sc := baseTest(t, sql, sc, 3)

		p.AddForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
			schema.AddForeignKeyConstraintOptions{
				Column:     "user_id",
				PrimaryKey: "ref",
			})

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, schema.Table("articles", sc), "fk_articles_authors")
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

		p, r, sc := baseTest(t, sql, sc, 4)

		p.AddForeignKeyConstraint(schema.Table("orders", sc), schema.Table("carts", sc),
			schema.AddForeignKeyConstraintOptions{
				CompositePrimaryKey: []string{"shop_id", "user_id"},
			})

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, schema.Table("orders", sc), "fk_orders_carts")
	})

	t.Run("with on delete", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 5)

		p.AddForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
			schema.AddForeignKeyConstraintOptions{
				OnDelete: "cascade",
			})

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, schema.Table("articles", sc), "fk_articles_authors")
	})

	t.Run("with on update", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 6)

		p.AddForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
			schema.AddForeignKeyConstraintOptions{
				OnUpdate: "cascade",
			})

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, schema.Table("articles", sc), "fk_articles_authors")
	})

	t.Run("with deferrable", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 7)

		p.AddForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
			schema.AddForeignKeyConstraintOptions{
				Deferrable: "DEFERRABLE INITIALLY DEFERRED",
			})

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, schema.Table("articles", sc), "fk_articles_authors")
	})

	t.Run("with no validate", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 8)

		p.AddForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
			schema.AddForeignKeyConstraintOptions{
				Validate: utils.Ptr(false),
			})

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintExist(t, p, schema.Table("articles", sc), "fk_articles_authors")
	})

	t.Run("with / without IfNotExists", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, base, sc, 9)

		p.AddForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
			schema.AddForeignKeyConstraintOptions{})

		require.Panics(t, func() {
			p.AddForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
				schema.AddForeignKeyConstraintOptions{})
		})

		t.Run("ensure no panic if param IfNotExists is true", func(t *testing.T) {
			p.AddForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
				schema.AddForeignKeyConstraintOptions{IfNotExists: true})
		})
	})
}

func TestPostgres_DropForeignKeyConstraint(t *testing.T) {
	t.Parallel()
	sc := "tst_pg_drop_foreign_key_constraint"

	base := `
create table {schema}.articles( id serial PRIMARY KEY, name text, author_id integer); 
create table {schema}.authors( id serial PRIMARY KEY, name text);
alter table {schema}.articles add constraint fk_articles_authors foreign key (author_id) references {schema}.authors(id);`

	t.Run("nominal", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 0)

		p.DropForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc))

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintNotExist(t, p, schema.Table("articles", sc), "fk_articles_authors")
	})

	t.Run("with custom name", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 1)

		p.DropForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
			schema.DropForeignKeyConstraintOptions{
				ForeignKeyName: "fk_articles_authors",
			})

		testutils.AssertSnapshotDiff(t, r.String())

		assertConstraintNotExist(t, p, schema.Table("articles", sc), "fk_articles_authors")
	})

	t.Run("with / without IfExists", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, base, sc, 2)

		p.DropForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc))
		require.Panics(t, func() {
			p.DropForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc))
		})

		t.Run("ensure no panic if param IfExists is true", func(t *testing.T) {
			p.DropForeignKeyConstraint(schema.Table("articles", sc), schema.Table("authors", sc),
				schema.DropForeignKeyConstraintOptions{IfExists: true})
		})
	})
}

func TestPostgres_AddPrimaryKeyConstraint(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_add_primary_key_constraint"

	base := `create table {schema}.articles (id integer, name text)`

	t.Run("simple primary key", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 0)

		p.AddPrimaryKeyConstraint(schema.Table("articles", sc), []string{"id"}, schema.PrimaryKeyConstraintOptions{})

		testutils.AssertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "id", DataType: "integer", PrimaryKey: true},
			{ColumnName: "name", DataType: "text"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("composite primary key", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 1)

		p.AddPrimaryKeyConstraint(schema.Table("articles", sc), []string{"id", "name"},
			schema.PrimaryKeyConstraintOptions{})

		testutils.AssertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "id", DataType: "integer", PrimaryKey: true},
			{ColumnName: "name", DataType: "text", PrimaryKey: true},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("if not exists", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, base, sc, 2)

		p.AddPrimaryKeyConstraint(schema.Table("articles", sc), []string{"id"}, schema.PrimaryKeyConstraintOptions{})
		require.Panics(t, func() {
			p.AddPrimaryKeyConstraint(schema.Table("articles", sc), []string{"id"},
				schema.PrimaryKeyConstraintOptions{})
		})
		p.AddPrimaryKeyConstraint(schema.Table("articles", sc), []string{"id"},
			schema.PrimaryKeyConstraintOptions{IfNotExists: true})
	})
}

func TestPostgres_DropPrimaryKeyConstraint(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_drop_primary_key_constraint"

	base := `
	create table {schema}.articles(id integer, name text, PRIMARY KEY (id))`

	t.Run("simple primary key", func(t *testing.T) {
		t.Parallel()
		p, r, sc := baseTest(t, base, sc, 0)

		p.DropPrimaryKeyConstraint(schema.Table("articles", sc))

		testutils.AssertSnapshotDiff(t, r.String())

		require.Equal(t, []columnInfo{
			{ColumnName: "id", DataType: "integer"},
			{ColumnName: "name", DataType: "text"},
		}, dumpColumns(t, p, schema.Table("articles", sc)))
	})

	t.Run("if exists", func(t *testing.T) {
		t.Parallel()
		p, _, sc := baseTest(t, base, sc, 1)

		p.DropPrimaryKeyConstraint(schema.Table("articles", sc))
		require.Panics(t, func() {
			p.DropPrimaryKeyConstraint(schema.Table("articles", sc))
		})

		p.DropPrimaryKeyConstraint(schema.Table("articles", sc), schema.DropPrimaryKeyConstraintOptions{IfExists: true})
	})
}
