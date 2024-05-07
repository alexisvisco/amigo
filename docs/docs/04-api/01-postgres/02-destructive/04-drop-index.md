# Drop index

The `DropIndex` function allows you to drop a specified index from a table in the database. It accepts the name of the table and a list of column names that make up the index, as well as optional parameters for configuring the drop operation.

#### Basic Usage

The function follows this format:

```go
p.DropIndex(table, columns opt)
```

- `table`: The name of the table from which you want to drop the index.
- `columns`: A list of column names that make up the index.
- `opt`: Optional parameters for configuring the drop operation.

#### Defining Drop Index Options

The function accepts optional `schema.DropIndexOptions` to customize the drop operation:

- **Index Name**:
    - `IndexName`: Specify the name of the index you want to drop. If not specified, the default index name is generated based on the table and columns.

- **If Exists**:
    - `IfExists`: If set to `true`, the function will drop the index only if it exists. This avoids errors when trying to drop an index that does not exist.

- **Reversible**:
    - `Reversible`: Provides options to reverse the drop operation by recreating the index. Specify `schema.IndexOptions` to define how the index will be recreated.

#### Examples

- **Dropping an index from a table**:

    ```go
    p.DropIndex("products", []string{"name"}, schema.DropIndexOptions{})
    ```

  The SQL generated would be:

    ```sql
    DROP INDEX idx_products_name
    ```

- **Dropping an index if it exists**:

    ```go
    p.DropIndex("products", []string{"name"}, schema.DropIndexOptions{IfExists: true})
    ```

  The SQL generated would be:

    ```sql
    DROP INDEX IF EXISTS idx_products_name
    ```

- **Making the drop operation reversible**:

  You can specify `Reversible` options to allow the operation to be reversed by recreating the index.

  If you do not specify the reversible options, the index will be dropped without the possibility of recreating it.

    ```go
    p.DropIndex("products", []string{"name"}, schema.DropIndexOptions{
        Reversible: &schema.IndexOptions{},
    })
    ```

  The SQL generated would be:

    ```sql
    CREATE INDEX idx_products_name ON "products" (name)
    ```
