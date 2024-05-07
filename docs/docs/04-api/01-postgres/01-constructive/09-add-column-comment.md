# Add column comment

The `AddColumnComment` function allows you to add a comment to a specified column in a database table. It accepts the table name, column name, the comment, and optional options for the column comment.

#### Basic Usage

The function follows this format:

```go
p.AddColumnComment(tableName, columnName, comment, opt)
```

- `tableName`: The name of the table containing the column to which you want to add the comment.
- `columnName`: The name of the column to which you want to add the comment.
- `comment`: A pointer to a string representing the comment you want to add. Pass `nil` to remove the comment.
- `opt`: Optional parameters for configuring the column comment.

#### Defining Column Comment Options

The function accepts optional `schema.ColumnCommentOptions` to customize the column comment:

- **Comment**:
    - `Comment`: A pointer to the comment text to be added to the column.

- **Reversible**:
    - `Reversible`: This option enables the migrator to reverse the operation.

#### Examples

- **Adding a comment to a column**:

    ```go
    p.AddColumnComment("users", "name", utils.Ptr("The name of the User"))
    ```

  The SQL generated would be:

    ```sql
    COMMENT ON COLUMN "users"."name" IS 'The name of the User'
    ```

- **Removing a comment from a column**:

    ```go
    p.AddColumnComment("users", "name", nil)
    ```

  The SQL generated would be:

    ```sql
    COMMENT ON COLUMN "users"."name" IS NULL
    ```

#### Using Reversible Option

The `Reversible` option allows the operation to be reversed. Provide a `schema.ColumnCommentOptions` struct as a reversible option to specify how to reverse the operation.
