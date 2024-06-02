# Destructive operations

They are the operations that drop tables, columns, indexes, constraints, and so on.

- [DropIndex(table schema.TableName, columns []string, opts ...schema.DropIndexOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/sqlite#Schema.DropIndex)

Usually you will use these functions in the `down` function of a migration, but you can use them in the `up` function too.
If you want to have the reverse operation of a destructive operation, you can use the `reversible` options. 

Example: 

```go
p.DropTable("users", schema.DropTableOptions{
	Reversible: &TableOption{
        schema.TableName: "users",
        PostgresTableDefinition: Innerschema.Tablefunc(t *PostgresTableDef) {
            t.Serial("id")
            t.String("name")
        }),
}})
```

In that case, if you are in a `change` function in your migration, the library will at the up operation drop the table `users` and at the down operation re-create the table `users` with the columns `id` and `name`.
