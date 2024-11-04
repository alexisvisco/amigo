# Working with existing schema

If you have an existing schema and you want to use mig to manage it, you can do it by following these steps:

```
mig create origin_schema --dump --skip
```

This command will create a new migration file `<version>_origin_schema.go` in the `migrations` folder.

This flag will dump the schema of your database with `pg_dump` (make sure you have it). 

The `--skip` flag will insert the current version of the schema into the `mig_schema_versions` table without running the migration (because the schema already exists).

You can specify some flags: 
- `--dump-schema` to specify the schema to dump. (default is `public`)
- `--pg-dump-path` to specify the path to the `pg_dump` command.
- `--shell-path` to specify the path to the shell command. (default is `/bin/bash`)
