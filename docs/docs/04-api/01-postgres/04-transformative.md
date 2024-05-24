# Transformative operations

They are the operations that change the data in the database.

- [RenameColumn(tableName schema.TableName, oldColumnName, newColumnName string)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.RenameColumn)

- [RenameTable(tableName schema.TableName, newTableName string)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.RenameTable)

- [ChangeColumnType(tableName schema.TableName, columnName string, columnType schema.ColumnType, opts ...schema.ChangeColumnTypeOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.ChangeColumnType)

- [ChangeColumnDefault(tableName schema.TableName, columnName string, defaultValue string, opts ...schema.ChangeColumnDefaultOptions)](https://pkg.go.dev/github.com/alexisvisco/amigo/pkg/schema/pg#Schema.ChangeColumnDefault)
