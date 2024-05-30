# Informative operations

They are the operations that give you information about the database schema.


- [TableExist(tableName schema.TableName) bool](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.TableExist)

- [ColumnExist(tableName schema.TableName, columnName string) bool](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.ColumnExist)

- [ConstraintExist(tableName schema.TableName, constraintName string) bool](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.ConstraintExist)

- [IndexExist(tableName schema.TableName, indexName string) bool](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.IndexExist)

- [PrimaryKeyExist(tableName schema.TableName) bool](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.PrimaryKeyExist)

- [FindAppliedVersions() []string](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/base#Schema.FindAppliedVersions)

- [FindEnumUsage(name string, schemaName *string) []schema.EnumUsage](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.FindEnumUsage)

- [ListEnumValues(name string, schemaName *string) []string](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.ListEnumValues)

These functions are not reversible.
