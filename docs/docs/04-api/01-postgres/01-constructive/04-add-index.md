# Add index

The `AddIndexConstraint` function allows you to add a new index to an existing table in the database. It accepts the table name, a list of column names to index, and index options.

#### Basic Usage

The function follows this format:

```go
p.AddIndexConstraint(tableName, columns, opt)
```

- `tableName`: The name of the table to which you want to add the index.
- `columns`: A list of column names to index.
- `opt`: Options for configuring the index creation.

#### Defining Index Options

The function accepts `schema.IndexOptions` to customize the index creation:

- **Index Naming**:
    - `IndexNameBuilder`: A function that builds the name of the index. If nil, a default name is used.
    - `IndexName`: An explicit name for the index.

- **Existence Check**:
    - `IfNotExists`: If set to `true`, the function will add the index only if it does not already exist in the table.

- **Unique**:
    - `Unique`: If set to `true`, the index will be unique.

- **Concurrent**:
    - `Concurrent`: If set to `true`, the index will be created concurrently.

- **Method**:
    - `Method`: Specifies the index method.

- **Order**:
    - `Order`: Specifies the order of the index (e.g., `ASC` or `DESC`).

- **Order Per Column**:
    - `OrderPerColumn`: Specifies the order of the index per column.

- **Predicate**:
    - `Predicate`: Specifies a predicate for the index.

#### Examples

- **Creating a simple index**:

    ```go
    p.AddIndexConstraint("products", []string{"name"})
    ```

  The SQL generated would be:

    ```sql
    CREATE INDEX idx_products_name ON "products" (name)
    ```

- **Creating a unique index**:

    ```go
    p.AddIndexConstraint("products", []string{"name"}, IndexOptions{Unique: true})
    ```

  The SQL generated would be:

    ```sql
    CREATE UNIQUE INDEX idx_products_name ON "products" (name)
    ```

- **Creating an index with a custom name**:

    ```go
    p.AddIndexConstraint("products", []string{"name"}, IndexOptions{IndexNameBuilder: func(table schema.TableName, columns []string) string {
        return "index_products_on_name"
    }})
    ```

  The SQL generated would be:

    ```sql
    CREATE INDEX index_products_on_name ON "products" (name)
    ```

- **Creating an index if it does not exist**:

    ```go
    p.AddIndexConstraint("products", []string{"name"}, IndexOptions{IfNotExists: true})
    ```

  The SQL generated would be:

    ```sql
    CREATE INDEX IF NOT EXISTS idx_products_name ON "products" (name)
    ```

- **Creating an index with a method**:

    ```go
    p.AddIndexConstraint("products", []string{"name"}, IndexOptions{Method: "btree"})
    ```

  The SQL generated would be:

    ```sql
    CREATE INDEX idx_products_name ON "products" USING btree (name)
    ```

- **Creating an index concurrently**:

    ```go
    p.AddIndexConstraint("products", []string{"name"}, IndexOptions{Concurrent: true})
    ```

  The SQL generated would be:

    ```sql
    CREATE INDEX CONCURRENTLY idx_products_name ON "products" (name)
    ```

- **Creating an index with a custom order**:

    ```go
    p.AddIndexConstraint("products", []string{"name"}, IndexOptions{Order: "DESC"})
    ```

  The SQL generated would be:

    ```sql
    CREATE INDEX idx_products_name ON "products" (name DESC)
    ```

- **Creating an index with a custom order per column**:

    ```go
    p.AddIndexConstraint("products", []string{"name", "price"}, IndexOptions{OrderPerColumn: map[string]string{"name": "DESC"}})
    ```

  The SQL generated would be:

    ```sql
    CREATE INDEX idx_products_name_price ON "products" (name DESC, price)
    ```

- **Creating an index with a predicate**:

    ```go
    p.AddIndexConstraint("products", []string{"name"}, IndexOptions{Predicate: "name IS NOT NULL"})
    ```

  The SQL generated would be:

    ```sql
    CREATE INDEX idx_products_name ON "products" (name) WHERE name IS NOT NULL
    ```
