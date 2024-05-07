# Drop Check constraint

The `DropCheckConstraint` function allows you to drop a specified check constraint from a table in the database. It accepts the table name, constraint name, and optional parameters for configuring the drop operation.

#### Basic Usage

The function follows this format:

```go
p.DropCheckConstraint(tableName, constraintName, opt)
```

- `tableName`: The name of the table from which you want to drop the check constraint.
- `constraintName`: The name of the check constraint you want to drop.
- `opt`: Optional parameters for configuring the drop operation.

#### Defining Drop Check Constraint Options

The function accepts optional `schema.DropCheckConstraintOptions` to customize the drop operation:

- **If Exists**:
    - `IfExists`: If set to `true`, the function will drop the constraint only if it exists. This avoids errors when trying to drop a constraint that does not exist.

- **Reversible**:
    - `Reversible`: Provides options to reverse the drop operation by creating the constraint. Specify `schema.CheckConstraintOptions` to define how the constraint will be recreated.

#### Examples

- **Dropping a check constraint from a table**:

    ```go
    p.DropCheckConstraint("products", "price_check")
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "products" DROP CONSTRAINT price_check
    ```

- **Dropping a check constraint if it exists**:

    ```go
    p.DropCheckConstraint("products", "price_check", schema.DropCheckConstraintOptions{IfExists: true})
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "products" DROP CONSTRAINT IF EXISTS price_check
    ```

- **Making the drop operation reversible**:

  You can specify `Reversible` options to allow the operation to be reversed by recreating the check constraint.

  If you do not specify the reversible options, the check constraint will be dropped without the possibility of recreating it.

    ```go
    p.DropCheckConstraint("products", "price_check", schema.DropCheckConstraintOptions{
        Reversible: schema.CheckConstraintOptions{
            Expression: "price > 0",
        },
    })
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "products" ADD CONSTRAINT price_check CHECK (price > 0)
    ```

