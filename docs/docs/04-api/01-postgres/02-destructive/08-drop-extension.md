# Drop extension

The `DropExtension` function allows you to drop a specified extension from the database. It accepts the name of the extension you want to drop and optional parameters for configuring the drop operation.

#### Basic Usage

The function follows this format:

```go
p.DropExtension(name, opt)
```

- `name`: The name of the extension you want to drop.
- `opt`: Optional parameters for configuring the drop operation.

#### Defining Drop Extension Options

The function accepts optional `schema.DropExtensionOptions` to customize the drop operation:

- **If Exists**:
    - `IfExists`: If set to `true`, the function will drop the extension only if it exists. This avoids errors when trying to drop an extension that does not exist.

- **Reversible**:
    - `Reversible`: Provides options to reverse the drop operation by recreating the extension. Specify `schema.ExtensionOptions` to define how the extension will be recreated.

#### Examples

- **Dropping an extension from the database**:

    ```go
    p.DropExtension("uuid", DropExtensionOptions{})
    ```

  The SQL generated would be:

    ```sql
    DROP EXTENSION IF EXISTS "uuid-ossp"
    ```

- **Dropping an extension if it exists**:

    ```go
    p.DropExtension("uuid", DropExtensionOptions{IfExists: true})
    ```

  The SQL generated would be:

    ```sql
    DROP EXTENSION IF EXISTS "uuid-ossp"
    ```

- **Making the drop operation reversible**:

  You can specify `Reversible` options to allow the operation to be reversed by recreating the extension.
  
  If you do not specify the reversible options, the extension will be dropped without the possibility of recreating it.

    ```go
    p.DropExtension("uuid", DropExtensionOptions{
        Reversible: &schema.ExtensionOptions{},
    })
    ```

  The SQL generated would be:

    ```sql
    CREATE EXTENSION "uuid-ossp"
    ```

