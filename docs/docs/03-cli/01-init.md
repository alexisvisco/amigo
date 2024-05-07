# Init

The `init` command allow you to boilerplate some files: 
- A `migrations` folder where you will write your migrations.
- A `.mig` folder where mig stores its configuration and the main file to run migrations.
- A migration file to setup the table that will store the migration versions.

## Flags
- `--mig-folder` is the folder where mig stores its configuration and the main file to run migrations. Default is `.mig`.
- `--package` is the package name of the migrations. Default is `migrations`.
- `--folder` is the folder where you will write your migrations. Default is `migrations`.
- `--schema-version-table` is the table that will store the migration versions. Default is `public.mig_schema_versions`.

## Note

When you set a flags with the `mig init` command, it can be useful to add them in the `context` command to avoid passing them each time you run a command.
