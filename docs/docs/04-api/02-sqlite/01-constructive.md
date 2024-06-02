# Constructive operations

They are the operations that create, alter, or drop tables, columns, indexes, constraints, and so on.

- [AddIndex(table schema.TableName, columns []string, option ...schema.IndexOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/sqlite#Schema.AddIndex)


Each of this functions are reversible, it means that in a migration that implement the `change` function, when you
rollback the migration you don't have to write manually the rollback operation, the library will do it for you.
