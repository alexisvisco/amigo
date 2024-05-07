# Add column

The `AddColumn` function allows you to add a new column to an existing table in the database. It accepts the table name,
the column name, the column type, and optional column options.

#### Basic Usage

The function follows this format:

```go
p.AddColumn(tableName, columnName, columnType, opt)
```

- `tableName`: The name of the table to which you want to add the new column.
- `columnName`: The name of the new column.
- `columnType`: The type of the new column. You can use the predefined types provided by `schema.ColumnType`.
- `opt`: Optional parameters for configuring the column addition.

#### Defining Column Options

The function accepts optional `schema.ColumnOptions` to customize the column:

- **Type-specific Options**:
    - `Precision`: The precision of the column for types such
      as `ColumnTypeDecimal`, `ColumnTypeNumeric`, `ColumnTypeDatetime`, and `ColumnTypeTime`.
    - `Scale`: The scale of the column for types such as `ColumnTypeDecimal` and `ColumnTypeNumeric`.

- **Default Value**:
    - `Default`: The default value of the column.

- **Constraints**:
    - `NotNull`: If set to `true`, the column will be defined with a `NOT NULL` constraint.
    - `PrimaryKey`: If set to `true`, the column will be defined as a primary key.

- **Limits and Arrays**:
    - `Limit`: The maximum length of the column. For example, the number of characters for `ColumnTypeString` and the
      number of bytes for `ColumnTypeText`, `ColumnTypeBinary`, `ColumnTypeBlob`, and `ColumnTypeInteger`.
    - `Array`: If set to `true`, the column will be defined as an array.

- **Existence Check**:
    - `IfNotExists`: If set to `true`, the function will add the column only if it does not already exist in the table.

- **Comment**:
    - `Comment`: Optional comment for the column.

Here is a guide on how to use the `AddColumn` function in your Go program with the provided code and an overview of the parameters and options for customizing column creation.

#### Examples

- **Adding a basic column**:

    ```go
    p.AddColumn("users", "picture", schema.ColumnTypeBinary)
    ```

  Generates:

    ```sql
    ALTER TABLE "users" ADD "picture" BYTEA
    ```

- **Adding a column with options**:

    ```go
    p.AddColumn("articles", "status", schema.ColumnTypeString, schema.ColumnOptions{
        Limit: 20,
        Default: "draft",
        NotNull: false,
    })
    ```

  Generates:

    ```sql
    ALTER TABLE "articles" ADD "status" VARCHAR(20) DEFAULT 'draft' NOT NULL
    ```

- **Adding a column with precision and scale**:

    ```go
    p.AddColumn("answers", "bill_gates_money", schema.ColumnTypeDecimal, schema.ColumnOptions{
        Precision: 15,
        Scale: 2,
    })
    ```

  Generates:

    ```sql
    ALTER TABLE "answers" ADD "bill_gates_money" DECIMAL(15,2)
    ```

- **Adding an array column**:

    ```go
    p.AddColumn("users", "skills", schema.ColumnTypeText, schema.ColumnOptions{Array: true})
    ```

  Generates:

    ```sql
    ALTER TABLE "users" ADD "skills" TEXT[]
    ```

- **Adding a column with a custom type**:

    ```go
    p.AddColumn("shapes", "triangle", "polygon")
    ```

  Generates:

    ```sql
    ALTER TABLE "shapes" ADD "triangle" POLYGON
    ```

- **Adding a column if it does not exist**:

    ```go
    p.AddColumn("shapes", "triangle", "polygon", schema.ColumnOptions{IfNotExists: true})
    ```

  Generates:

    ```sql
    ALTER TABLE "shapes" ADD "triangle" IF NOT EXISTS POLYGON
    ```

