# Initialize mig

To start using mig, you need to initialize it. This process creates few things:
- A `migrations` folder where you will write your migrations.
- A `.amigo` folder where mig stores its configuration and the main file to run migrations.
- A migration file to setup the table that will store the migration versions.

To initialize mig, run the following command:

```sh
amigo init
```

You can also create a context to the repository to avoid passing flags each time you run a command. To do so, run the following command:


### Postgres:
```sh
amigo context --dsn "postgres://user:password@localhost:5432/dbname"
```

### SQLite:
```sh
amigo context --dsn "sqlite:/path/to/db.sqlite" --schema-version-table mig_schema_versions
```
Note: The `--schema-version-table` flag is optional and is used to specify the table where mig will store the migration versions. By default, mig uses `public.mig_schema_versions` but since SQLite does not support schemas, you can specify the table name.


### Configuration

A config.yml file will be created in the `.amigo` folder. You can edit it to add more configurations.

It contains the following fields:
```yaml
dsn: postgres://user:password@localhost:5432/dbname
folder: migrations
json: false
amigo-folder: .amigo
package: migrations
pg-dump-path: pg_dump
schema-version-table: public.mig_schema_versions
shell-path: /bin/bash
debug: false
show-sql: false
show-sql-highlighting: true
```
