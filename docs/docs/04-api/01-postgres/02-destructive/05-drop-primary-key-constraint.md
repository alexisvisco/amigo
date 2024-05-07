# Drop Primary Key Constraint

The `DropPrimaryKeyConstraint` function allows you to drop a primary key from a specified table in your database. It accepts the name of the table from which you want to drop the primary key and optional parameters for configuring the drop operation.

#### Basic Usage

The function follows this format:

```go
p.DropPrimaryKeyConstraint(tableName, opt)
```

- `tableName`: The name of the table from which you want to drop the primary key.
- `opt`: Optional parameters for configuring the drop operation.

#### Defining Drop Primary Key Constraint Options

The function accepts optional `schema.DropPrimaryKeyConstraintOptions` to customize the drop operation:

- **If Exists**:
    - `IfExists`: If set to `true`, the function will drop the primary key only if it exists. This avoids errors when trying to drop a primary key that does not exist.

- **Reversible**:
    - `Reversible`: Provides options to reverse the drop operation by recreating the primary key. Specify `schema.PrimaryKeyConstraintOptions` to define how the primary key will be recreated.

#### Examples

- **Dropping a primary key from the table**:

    ```go
    p.DropPrimaryKeyConstraint("users")
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "users" DROP CONSTRAINT pk_users
    ```

- **Dropping a primary key if it exists**:

    ```go
    p.DropPrimaryKeyConstraint("users", schema.DropPrimaryKeyConstraintOptions{IfExists: true})
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "users" DROP CONSTRAINT IF EXISTS pk_users
    ```

- **Making the drop operation reversible**:

  You can specify `Reversible` options to allow the operation to be reversed by recreating the primary key.

  If you do not specify the reversible options, the primary key will be dropped without the possibility of recreating it.

    ```go
    p.DropPrimaryKeyConstraint("users", schema.DropPrimaryKeyConstraintOptions{
        Reversible: schema.PrimaryKeyConstraintOptions{Columns: []string{"id"}},
    })
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "users" ADD CONSTRAINT PRIMARY KEY (id)
    ```

