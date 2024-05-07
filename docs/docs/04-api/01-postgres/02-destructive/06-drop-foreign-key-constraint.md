# Drop Foreign Key constraint

The `DropForeignKeyConstraint` function allows you to drop a specified foreign key constraint from a table in the database. It accepts the names of the source and target tables (or the associated foreign key), as well as optional parameters for configuring the drop operation.

#### Basic Usage

The function follows this format:

```go
p.DropForeignKeyConstraint(from, to, opt)
```

- `from`: The name of the source table containing the foreign key you want to drop.
- `to`: The name of the target table to which the foreign key points.
- `opt`: Optional parameters for configuring the drop operation.

#### Defining Drop Foreign Key Constraint Options

The function accepts optional `schema.DropForeignKeyConstraintOptions` to customize the drop operation:

- **If Exists**:
    - `IfExists`: If set to `true`, the function will drop the constraint only if it exists. This avoids errors when trying to drop a constraint that does not exist.

- **Reversible**:
    - `Reversible`: Provides options to reverse the drop operation by creating the constraint. Specify `schema.AddForeignKeyConstraintOptions` to define how the constraint will be recreated.

#### Examples

- **Dropping a foreign key from a table**:

    ```go
    p.DropForeignKeyConstraint("articles", "authors")
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "articles" DROP CONSTRAINT fk_articles_authors
    ```

- **Dropping a foreign key if it exists**:

    ```go
    p.DropForeignKeyConstraint("articles", "authors", schema.DropForeignKeyConstraintOptions{IfExists: true})
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "articles" DROP CONSTRAINT IF EXISTS fk_articles_authors
    ```

- **Making the drop operation reversible**:

  You can specify `Reversible` options to allow the operation to be reversed by recreating the foreign key constraint.
  
  If you do not specify the reversible options, the foreign key constraint will be dropped without the possibility of recreating it.

    ```go
    p.DropForeignKeyConstraint("articles", "authors", schema.DropForeignKeyConstraintOptions{
        Reversible: schema.AddForeignKeyConstraintOptions{
            Column: "author_id",
            PrimaryKey: "lng_id",
        },
    })
    ```

  The SQL generated would be:

    ```sql
    ALTER TABLE "articles" ADD CONSTRAINT fk_articles_authors FOREIGN KEY (author_id) REFERENCES "authors" (lng_id)
    ```

