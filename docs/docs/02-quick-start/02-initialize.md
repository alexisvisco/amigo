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

```sh
amigo context --dsn "postgres://user:password@localhost:5432/dbname"
```

A config.yml file will be created in the `.mig` folder. You can edit it to add more configurations.

It contains the following fields:
```yaml
dsn: postgres://user:password@localhost:5432/dbname
folder: migrations
json: false
mig-folder: .amigo
package: migrations
pg-dump-path: pg_dump
schema-version-table: public.mig_schema_versions
shell-path: /bin/bash
debug: false
```
