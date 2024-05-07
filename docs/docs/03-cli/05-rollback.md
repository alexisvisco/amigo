# Rollback

The `rollback` command allows you to rollback the last migration.

To rollback the last migration, run the following command:

```sh
mig rollback
```

## Flags
- `--dry-run` will run the migrations without applying them.
- `--timeout` is the timeout for the migration (default is 2m0s).
- `--version` will rollback a specific version. The format is `20240502083700` or `20240502083700_name.go`.
- `--steps` will rollback the last `n` migrations. (default is 1)
- `--continue-on-error` will not rollback the migration if an error occurs.
