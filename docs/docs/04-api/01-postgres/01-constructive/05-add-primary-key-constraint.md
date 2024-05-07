# Add primary key

The `AddPrimaryKeyConstraint` function allows you to add a new primary key constraint to an existing table in the database. It accepts the table name, a list of column names that make up the primary key, and optional primary key constraint options.

#### Basic Usage

The function follows this format:

```go
p.AddPrimaryKeyConstraint(tableName, columns, opt)
```

- `tableName`: The name of the table to which you want to add the primary key constraint.
- `columns`: A list of column names that make up the primary key.
- `opt`: Optional parameters for configuring the primary key constraint.

#### Defining Primary Key Constraint Options

The function accepts optional `schema.PrimaryKeyConstraintOptions` to customize the primary key constraint:

- **Existence Check**:
    - `IfNotExists`: If set to `true`, the function will add the primary key constraint only if it does not already exist in the table.

#### Examples

- **Adding a simple primary key constraint**:

    ```go
    p.AddPrimaryKeyConstraint("users", []string{"id"}, PrimaryKeyConstraintOptions{})
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "users" ADD CONSTRAINT PRIMARY KEY (id)
    ```

- **Adding a composite primary key**:

    ```go
    p.AddPrimaryKeyConstraint("users", []string{"id", "name"}, PrimaryKeyConstraintOptions{})
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "users" ADD CONSTRAINT PRIMARY KEY (id, name)
    ```

- **Adding a primary key if it does not exist**:

    ```go
    p.AddPrimaryKeyConstraint("users", []string{"id"}, PrimaryKeyConstraintOptions{IfNotExists: true})
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "users" ADD CONSTRAINT IF NOT EXISTS PRIMARY KEY (id)
    ```

