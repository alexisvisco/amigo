# Destructive operations

They are the operations that drop tables, columns, indexes, constraints, and so on.


- [DropCheckConstraint(tableName schema.TableName, constraintName string, opts ...schema.DropCheckConstraintOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.DropCheckConstraint)

- [DropColumn(tableName schema.TableName, columnName string, opts ...schema.DropColumnOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.DropColumn)

- [DropExtension(name string, opts ...schema.DropExtensionOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.DropExtension)

- [DropForeignKeyConstraint(fromTable, toTable schema.TableName, opts ...schema.DropForeignKeyConstraintOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.DropForeignKeyConstraint)

- [DropIndex(table schema.TableName, columns []string, opts ...schema.DropIndexOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.DropIndex)

- [DropPrimaryKeyConstraint(tableName schema.TableName, opts ...schema.DropPrimaryKeyConstraintOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.DropPrimaryKeyConstraint)

- [DropTable(tableName schema.TableName, opts ...schema.DropTableOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.DropTable)

- [RemoveVersion(version string)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/base#Schema.RemoveVersion)

- [RenameColumn(tableName schema.TableName, oldColumnName, newColumnName string)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.RenameColumn)

- [DropEnum(name string, opts ...schema.DropEnumOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.DropEnum)

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
