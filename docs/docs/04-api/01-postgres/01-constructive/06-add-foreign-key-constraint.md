# Add foreign key

The `AddForeignKeyConstraint` function allows you to add a new foreign key constraint between two tables in the database. It accepts the source table (`fromTable`), the target table (`toTable`), and optional foreign key constraint options.

#### Basic Usage

The function follows this format:

```go
p.AddForeignKeyConstraint(fromTable, toTable, opt)
```

- `fromTable`: The source table containing the key column.
- `toTable`: The target table containing the referenced primary key.
- `opt`: Optional parameters for configuring the foreign key constraint.

#### Defining Foreign Key Constraint Options

The function accepts optional `schema.AddForeignKeyConstraintOptions` to customize the foreign key constraint:

- **Foreign Key Naming**:
    - `ForeignKeyName`: An explicit name for the foreign key constraint.
    - `ForeignKeyNameBuilder`: A function that builds the name of the foreign key constraint. If nil, a default name is used.

- **Column and Primary Key**:
    - `Column`: The foreign key column name in the source table. Defaults to `toTable.Singularize + "_id"`.
    - `PrimaryKey`: The primary key column name in the target table. Defaults to `id`.

- **Composite Primary Key**:
    - `CompositePrimaryKey`: A list of primary key column names in the target table.

- **On Delete and On Update Actions**:
    - `OnDelete`: Specifies the action to take on deletion (e.g., `"nullify"`, `"cascade"`, `"restrict"`).
    - `OnUpdate`: Specifies the action to take on update (e.g., `"nullify"`, `"cascade"`, `"restrict"`).

- **Existence Check**:
    - `IfNotExists`: If set to `true`, the function will add the foreign key constraint only if it does not already exist in the table.

- **Postgres Specific**:
    - `Validate`: Specifies whether the constraint should be validated. Defaults to `true`.
    - `Deferrable`: Specifies whether the foreign key should be deferrable.

#### Examples

- **Creating a simple foreign key**:

    ```go
    p.AddForeignKeyConstraint("articles", "authors", AddForeignKeyConstraintOptions{})
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "articles" ADD CONSTRAINT fk_articles_authors FOREIGN KEY (author_id) REFERENCES "authors" (id)
    ```

- **Creating a foreign key on a specific column**:

    ```go
    p.AddForeignKeyConstraint("articles", "users", AddForeignKeyConstraintOptions{Column: "author_id", PrimaryKey: "lng_id"})
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "articles" ADD CONSTRAINT fk_articles_users FOREIGN KEY (author_id) REFERENCES "users" (lng_id)
    ```

- **Creating a composite foreign key**:

  Assuming the `carts` table has a primary key on `(shop_id, user_id)`.

    ```go
    p.AddForeignKeyConstraint("orders", "carts", AddForeignKeyConstraintOptions{PrimaryKey: []string{"shop_id", "user_id"}})
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "orders" ADD CONSTRAINT fk_orders_carts FOREIGN KEY (cart_shop_id, cart_user_id) REFERENCES "carts" (shop_id, user_id)
    ```

- **Creating a cascading foreign key**:

    ```go
    p.AddForeignKeyConstraint("articles", "authors", AddForeignKeyConstraintOptions{OnDelete: "cascade"})
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "articles" ADD CONSTRAINT fk_articles_authors FOREIGN KEY (author_id) REFERENCES "authors" (id) ON DELETE CASCADE
    ```

