# Status

The `status` command will show the current status of the migrations.

It show the 10 most recent migrations and their status between applied and not applied.

The last line of the output will be the most recent migration.


```sh
amigo status

(20240530063939) create_table_schema_version   applied
(20240524090434) create_user_table             not applied
```
