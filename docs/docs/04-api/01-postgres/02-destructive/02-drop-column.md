# Drop column

The `DropColumn` function allows you to drop a specified column from a table in the database. It accepts the name of the table, the name of the column, and optional parameters for configuring the drop operation.

#### Basic Usage

The function follows this format:

```go
p.DropColumn(tableName, columnName, opt)
```

- `tableName`: The name of the table from which you want to drop the column.
- `columnName`: The name of the column you want to drop.
- `opt`: Optional parameters for configuring the drop operation.

#### Defining Drop Column Options

The function accepts optional `schema.DropColumnOptions` to customize the drop operation:

- **If Exists**:
    - `IfExists`: If set to `true`, the function will drop the column only if it exists. This avoids errors when trying to drop a column that does not exist.

- **Reversible**:
    - `Reversible`: Provides options to reverse the drop operation by creating the column. Specify `schema.ColumnOptions` to define how the column will be recreated.

#### Examples

- **Dropping a column from a table**:

    ```go
    p.DropColumn("users", "name")
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "users" DROP COLUMN "name"
    ```

- **Dropping a column if it exists**:

    ```go
    p.DropColumn("users", "name", schema.DropColumnOptions{IfExists: true})
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "users" DROP COLUMN IF EXISTS "name"
    ```

- **Making the drop operation reversible**:

  You can specify `Reversible` options to allow the operation to be reversed by recreating the column.
    
  If you do not specify the reversible options, the column will be dropped without the possibility of recreating it.

    ```go
    p.DropColumn("users", "name", schema.DropColumnOptions{
        Reversible: &schema.ReversibleColumn{ColumnType: "VARCHAR(255)"},
    })
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "users" ADD "name" VARCHAR(255)
    ```

