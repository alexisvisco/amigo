# Add extension

The `AddExtension` function allows you to add a new extension to the database. It accepts the extension name and optional extension options.

#### Basic Usage

The function follows this format:

```go
p.AddExtension(name, opt)
```

- `name`: The name of the extension you want to add.
- `opt`: Optional parameters for configuring the extension creation.

#### Defining Extension Options

The function accepts optional `schema.ExtensionOptions` to customize the extension creation:

- **Extension Name**:
    - `ExtensionName`: Specifies the name of the extension. By default, it is the same as `name`.

- **Schema**:
    - `Schema`: Specifies the schema where the extension will be created.

- **Existence Check**:
    - `IfNotExists`: If set to `true`, the function will add the extension only if it does not already exist in the database.

#### Examples

- **Adding an extension**:

    ```go
    p.AddExtension("uuid", ExtensionOptions{})
    ```

  The SQL generated would be:

    ```sql
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp"
    ```

- **Adding an extension in a specific schema**:

    ```go
    p.AddExtension("uuid", ExtensionOptions{Schema: "public"})
    ```

  The SQL generated would be:

    ```sql
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA public
    ```

- **Adding an extension if it does not exist**:

    ```go
    p.AddExtension("uuid", ExtensionOptions{IfNotExists: true})
    ```

  The SQL generated would be:

    ```sql
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp"
    ```
