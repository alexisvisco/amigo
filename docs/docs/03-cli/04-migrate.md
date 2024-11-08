# Migrate

The `migrate` command allows you to apply the migrations.

To apply the migrations, run the following command:

```sh
amigo migrate

------> migrating: create_user_table version: 20240524110434
-- create_table(table: users, {columns: id, name, email, created_at, updated_at}, {pk: id})
-- add_index(table: users, name: idx_users_name, columns: [name])
------> version migrated: 20240524090434
```

## Flags
- `--dry-run` will run the migrations without applying them.
- `--timeout` is the timeout for the migration (default is 2m0s).
- `--version` will apply a specific version. The format is `20240502083700` or `20240502083700_name.go`.
- `--continue-on-error` will not rollback the migration if an error occurs.
- `-d` dump the schema after migrating

