# Context

The `context` command allow you to save root flags in a configuration file. This file will be used by mig to avoid passing flags each time you run a command.

Example: 
```sh
amigo context --dsn "postgres://user:password@localhost:5432/dbname" --folder mgs --package mgs
```

Will create a `config.yml` file in the `.mig` folder with the following content:
```yaml
dsn: postgres://user:password@localhost:5432/dbname
folder: mgs
json: false
mig-folder: .mig
package: mgs
pg-dump-path: pg_dump
schema-version-table: public.mig_schema_versions
shell-path: /bin/bash
debug: false
show-sql: false
```
