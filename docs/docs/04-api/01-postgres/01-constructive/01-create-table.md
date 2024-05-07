# Create table

The `CreateTable` function allows you to create a new table in the database. It accepts a table name, a function to define the table's schema, and optional table options.

#### Basic Usage

The function follows this format:

```go
p.CreateTable(tableName, func(t *pg.PostgresTableDef) {
    // Define the columns and options for the table
}, opt)
```

- `tableName`: The name of the table you want to create.
- `func(t *pg.PostgresTableDef)`: A function that defines the schema of the table. You can use the `t` parameter to specify column types, indices, foreign keys, and more.
- `opt`: Optional parameters for configuring the table creation.

#### Defining Table Schema

In the function that defines the table schema, you can specify various column types and other table settings using methods provided by `pg.PostgresTableDef`. Here are some common examples:

- **Columns**:
    - `t.String("name")`: Adds a `name` column of type `text`.
    - `t.Integer("age")`: Adds an `age` column of type `integer`.
    - `t.Serial("id")`: Adds an `id` column of type `serial` (auto-incrementing integer).

- **Indices**:
    - `t.Index([]string{"name"})`: Creates an index on the `name` column.

- **Foreign Keys**:
    - `t.ForeignKey("articles")`: Adds a foreign key constraint that references the `articles` table.

- **Timestamps**:
    - `t.Timestamps()`: Adds `created_at` and `updated_at` columns with `TIMESTAMP` type and default values.

#### Table Options

You can pass additional options to configure the table creation:

- `IfNotExists`: If set to `true`, the table is only created if it does not already exist.
- `PrimaryKeys`: A list of column names that should be set as the primary key.
- `WithoutPrimaryKey`: If set to `true`, the table is created without a primary key.
- `Option`: A string that specifies additional SQL options for the table creation.

#### Examples

- **Creating a table with primary key**:

    ```go
    p.CreateTable("users", func(t *pg.PostgresTableDef) {
        t.Serial("id")         // id column as serial (primary key)
        t.String("name")       // name column as text
        t.Integer("age")       // age column as integer
    })
    ```
  Generates:

    ```sql
    CREATE TABLE "users" ( "id" SERIAL PRIMARY KEY, "name" TEXT, "age" INTEGER )
    ```
- **Creating a table without primary key**:

    ```go
    p.CreateTable("users", func(t *pg.PostgresTableDef) {
        t.String("name")       // name column as text
    }, schema.TableOptions{WithoutPrimaryKey: true})
    ```

  Generates:

    ```sql
    CREATE TABLE "users" ( "name" TEXT )
    ```


- **Creating a table with a composite primary key**:

    ```go
    p.CreateTable("users", func(t *pg.PostgresTableDef) {
        t.String("name")       // name column as text
        t.Integer("age")       // age column as integer
    }, schema.TableOptions{PrimaryKeys: []string{"name", "age"}})
    ```

  Generates:

    ```sql
    CREATE TABLE "users" ( "name" TEXT, "age" INTEGER, PRIMARY KEY ("name", "age") )
    ```

- **Adding an index to a table**:

    ```go
    p.CreateTable("users", func(t *pg.PostgresTableDef) {
        t.String("name")       // name column as text
        t.Index([]string{"name"})
    })
    ```

  Generates:

    ```sql
    CREATE TABLE "users" ( "name" TEXT );
    CREATE INDEX "idx_users_name" ON "users" ("name")
    ```

- **Adding a foreign key constraint to a table**:

    ```go
    p.CreateTable("users", func(t *pg.PostgresTableDef) {
        t.String("name")        // name column as text
        t.Integer("article_id") // article_id column as integer
        t.ForeignKey("articles")
    })
    ```

  Generates:

    ```sql
    CREATE TABLE "users" ( "name" TEXT, "article_id" INTEGER );
    ALTER TABLE "users" ADD CONSTRAINT "fk_users_article_id" FOREIGN KEY ("article_id") REFERENCES "articles" ("id")
    ```
- **Adding a custom column type**:

    ```go
    p.CreateTable("users", func(t *pg.PostgresTableDef) {
        t.String("first_name")
        t.String("last_name")
        t.Column("name", "GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED")
    })
    ```

  Generates:

    ```sql
    CREATE TABLE "users" ( "first_name" TEXT, "last_name" TEXT, "name" GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED )
    ```
