# Add check constraint

The `AddCheckConstraint` function allows you to add a check constraint to an existing table in the database. A check constraint is a rule that must be satisfied for the data in a column or a set of columns. The function accepts the table name, constraint name, a boolean condition (expression), and optional constraint options.

#### Basic Usage

The function follows this format:

```go
p.AddCheckConstraint(tableName, constraintName, expression, opt)
```

- `tableName`: The name of the table to which you want to add the check constraint.
- `constraintName`: The name of the check constraint.
- `expression`: A string representation of the boolean condition (expression) that must be satisfied.
- `opt`: Optional parameters for configuring the check constraint.

#### Defining Constraint Options

The function accepts optional `schema.CheckConstraintOptions` to customize the check constraint:

- **Constraint Naming**:
    - `ConstraintNameBuilder`: A function that builds the name of the constraint. If nil, a default name is used.

- **Existence Check**:
    - `IfNotExists`: If set to `true`, the function will add the constraint only if it does not already exist in the table.

- **Validation**:
    - `Validate`: Specify whether the constraint should be validated. Defaults to `true`.

#### Examples

- **Adding a basic check constraint**:

    ```go
    p.AddCheckConstraint("products", "price_check", "price > 0")
    ```

  This adds a check constraint named `price_check` to the `products` table that ensures the `price` column value is greater than zero.

  The SQL generated would be:

    ```sql
    ALTER TABLE "products" ADD CONSTRAINT price_check CHECK (price > 0)
    ```

- **Adding a check constraint with options**:

    ```go
    p.AddCheckConstraint("orders", "quantity_check", "quantity > 0", schema.CheckConstraintOptions{
        IfNotExists: true,
    })
    ```

  In this example, a check constraint named `quantity_check` is added to the `orders` table that ensures the `quantity` column value is greater than zero. The `IfNotExists` option ensures the constraint is added only if it does not already exist in the table.

  The SQL generated would be:

    ```sql
    ALTER TABLE "orders" ADD CONSTRAINT quantity_check CHECK (quantity > 0)
    ```
