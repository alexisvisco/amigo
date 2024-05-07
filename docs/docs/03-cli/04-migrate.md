# Migrate

The `migrate` command allows you to apply the migrations.

To apply the migrations, run the following command:

```sh
mig migrate
```

## Flags
- `--dry-run` will run the migrations without applying them.
- `--timeout` is the timeout for the migration (default is 2m0s).
- `--version` will apply a specific version. The format is `20240502083700` or `20240502083700_name.go`.
- `--continue-on-error` will not rollback the migration if an error occurs.

