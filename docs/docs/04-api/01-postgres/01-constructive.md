# Constructive operations

They are the operations that create, alter, or drop tables, columns, indexes, constraints, and so on.

- [CreateTable(tableName schema.TableName, f func(*PostgresTableDef), opts ...schema.TableOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.CreateTable)

- [AddColumn(tableName schema.TableName, columnName string, columnType schema.ColumnType, opts ...schema.ColumnOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.AddColumn)

- [AddColumnComment(tableName schema.TableName, columnName string, comment *string, opts ...schema.ColumnCommentOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.AddColumnComment)

- [AddCheckConstraint(tableName schema.TableName, constraintName string, expression string, opts ...schema.CheckConstraintOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.AddCheckConstraint)

- [AddExtension(name string, option ...schema.ExtensionOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.AddExtension)

- [AddForeignKey(fromTable, toTable schema.TableName, opts ...schema.AddForeignKeyConstraintOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.AddForeignKeyConstraint)

- [AddIndexConstraint(table schema.TableName, columns []string, option ...schema.IndexOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.AddIndexConstraint)

- [AddPrimaryKeyConstraint(tableName schema.TableName, columns []string, opts ...schema.PrimaryKeyConstraintOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.AddPrimaryKeyConstraint)

- [AddVersion(version string)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.AddVersion)


Each of this functions are reversible, it means that in a migration that implement the `change` function, when you
rollback the migration you don't have to write manually the rollback operation, the library will do it for you.
