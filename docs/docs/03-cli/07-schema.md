# Schema

The `schema` command allow to dump the schema to `db/schema.sql` (default path)

Database supported: 
- postgres (via pg_dump)

## Flags 

- `--schema-db-dump-schema` will change the schema (default public) to dump
- `--schema-out-path` will change the path to output the file (db/schema.sql)