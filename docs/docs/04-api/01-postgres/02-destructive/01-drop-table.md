# Drop table

The `DropTable` function allows you to drop a table from the database. It accepts the name of the table to drop and optional parameters to configure the drop operation.

#### Basic Usage

The function follows this format:

```go
p.DropTable(tableName, opt)
```

- `tableName`: The name of the table you want to drop.
- `opt`: Optional parameters for configuring the drop operation.

#### Defining Drop Table Options

The function accepts optional `schema.DropTableOptions` to customize the drop table operation:

- **If Exists**:
    - `IfExists`: If set to `true`, the function will drop the table only if it exists. This avoids errors when trying to drop a table that does not exist.

- **Reversible**:
    - `Reversible`: Provides options to reverse the drop operation by creating the table. You can specify `schema.TableOptions` to define how the table will be recreated.

#### Examples

- **Dropping a table**:

    ```go
    p.DropTable("users", schema.DropTableOptions{})
    ```

  The SQL generated would be:

    ```sql
    DROP TABLE "users"
    ```

- **Dropping a table if it exists**:

    ```go
    p.DropTable("users", schema.DropTableOptions{IfExists: true})
    ```

  The SQL generated would be:

    ```sql
    DROP TABLE IF EXISTS "users"
    ```

- **Making the drop operation reversible**:

  You can specify `Reversible` options to allow the operation to be reversed by recreating the table.

  If you do not specify the reversible options, the table will be dropped without the possibility of recreating it.

    ```go
    p.DropTable("users", schema.DropTableOptions{
        Reversible: &TableOptions{
            schema.TableName: "users",
            PostgresTableDefinition: func(t *PostgresTableDef) {
                t.Serial("id")
                t.String("name")
            },
        },
    })
    ```
  

  The SQL generated would be the creation of the "users" table:

    ```sql
    CREATE TABLE "users" ( "id" serial PRIMARY KEY, "name" text )
    ```




