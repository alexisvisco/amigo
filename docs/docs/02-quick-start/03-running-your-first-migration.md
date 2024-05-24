# Running your first migration

Since you have initialized mig, you can now run your first migration.

Before that make sure to import the driver that amigo have added in the imports of the main file in `.amigo/main.go`.

By default for postgres it imports `github.com/jackc/pgx/v5/stdlib` but you can change it and it will works. 

Amigo is driver agnostic and works with the `database/sql` package.
`pgx` provide a support for the `database/sql` package and is a good choice for postgres, but you can use any driver that support the `database/sql` package.




When you have installed the driver on you project, run the migration, execute the following command:

```sh
amigo migrate
```
