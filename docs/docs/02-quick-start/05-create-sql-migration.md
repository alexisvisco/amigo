# SQL Migration

If you have a custom SQL to execute like materialized view, procedures and you want autocompletion and IDE support from 
SQL, you can use the `--type sql` flag.

```sh
amigo create my_migration --type sql
```

### Example of a `sql` migration file:

```sql
-- todo: write up migrations here
-- migrate:down
-- todo: write down migrations here
```

It act as a classic migration, you can write your SQL in the `-- todo: write up migrations here` section.

All code below the `-- migrate:down` comment will be used to rollback the migration.