package pg

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/alexisvisco/amigo/pkg/utils/orderedmap"
	"strings"
)

// CreateTable creates a new table in the database.
//
// Example:
//
//	p.CreateTable("users", func(t *pg.PostgresTableDef) {
//		t.Serial("id")
//		t.FormatRecords("name")
//		t.Integer("age")
//	})
//
// Generates:
//
//	CREATE TABLE "users" ( "id" serial PRIMARY KEY, "name" text, "age" integer )
//
// note: the primary key column must be defined in the table.
//
// To create a table without a primary key:
//
//	p.CreateTable("users", func(t *pg.PostgresTableDef) {
//		t.FormatRecords("name")
//	}, schema.TableOptions{ WithoutPrimaryKey: true })
//
// Generates:
//
//	CREATE TABLE "users" ( "name" text )
//
// To create a table with a composite primary key:
//
//	p.CreateTable("users", func(t *pg.PostgresTableDef) {
//		t.FormatRecords("name")
//		t.Integer("age")
//	}, schema.TableOptions{ PrimaryKeys: []string{"name", "age"} })
//
// Generates:
//
//	CREATE TABLE "users" ( "name" text, "age" integer )
//	ALTER TABLE "users" ADD PRIMARY KEY ("name", "age")
//
// note: You can use PrimaryKeys to specify the primary key name (without creating a composite primary key).
//
// To add index to the table:
//
//	p.CreateTable("users", func(t *pg.PostgresTableDef) {
//		t.FormatRecords("name")
//		t.Index([]string{"name"})
//	})
//
// Generates:
//
//	CREATE TABLE "users" ( "name" text )
//	CREATE INDEX idx_users_name ON "users" (name)
//
// To add foreign key to the table:
//
//	p.CreateTable("users", func(t *pg.PostgresTableDef) {
//		t.FormatRecords("name")
//		t.Integer("article_id")
//		t.ForeignKey("articles")
//	})
//
// Generates:
//
//	CREATE TABLE "users" ( "name" text, "article_id" integer )
//	ALTER TABLE "users" ADD CONSTRAINT fk_users_articles FOREIGN KEY (article_id) REFERENCES "articles" (id)
//
// To add created_at, updated_at Columns to the table:
//
//	p.CreateTable("users", func(t *pg.PostgresTableDef) {
//		t.FormatRecords("name")
//		t.Timestamps()
//	})
//
// Generates:
//
//	CREATE TABLE "users" ( "name" text, "created_at" TIMESTAMP(6) DEFAULT 'now()', "updated_at" TIMESTAMP(6) DEFAULT 'now()' )
func (p *Schema) CreateTable(tableName schema.TableName, f func(*PostgresTableDef), opts ...schema.TableOptions) {
	options := schema.TableOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Table = tableName

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		p.rollbackMode().DropTable(tableName, schema.DropTableOptions{IfExists: true})
		return
	}

	var td *PostgresTableDef
	if options.TableDefinition != nil {
		td = options.TableDefinition.(*PostgresTableDef)
	} else {
		td = p.buildInnerTable(tableName, f, &options)
	}

	options.TableDefinition = td

	q := `CREATE TABLE {if_not_exists} {table_name} (
		{inner_table}
    ) {table_options}`

	replacer := utils.Replacer{
		"if_not_exists": utils.StrFuncPredicate(options.IfNotExists, "IF NOT EXISTS"),
		"table_name":    utils.StrFunc(tableName.String()),
		"inner_table":   utils.StrFunc(strings.Join(td.innerTable, ",\n\t\t")),
		"table_options": utils.StrFuncPredicate(options.Option != "", options.Option),
	}

	_, err := p.DB.ExecContext(p.Context.Context, replacer.Replace(q))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while creating table: %w", err))
		return
	}

	p.Context.AddTableCreated(options)

	for _, afterCreate := range td.deferCreationAction {
		afterCreate()
	}
}

func (p *Schema) buildInnerTable(tableName schema.TableName, f func(*PostgresTableDef), options *schema.TableOptions) *PostgresTableDef {
	tableDef := &PostgresTableDef{
		parent:  p,
		table:   tableName,
		columns: orderedmap.New[*schema.ColumnOptions](),
	}

	f(tableDef)

	p.handlePrimaryKeysForCreateTable(tableName, options, tableDef)

	var innerTable []string

	for entry := range tableDef.columns.Iterate() {
		if entry.Value.PrimaryKey {
			entry.Value.NotNull = true
			entry.Value.Constraints = append(entry.Value.Constraints, schema.PrimaryKeyConstraintOptions{})
		}
		innerTable = append(innerTable, p.column(*entry.Value))
	}

	tableDef.innerTable = innerTable

	return tableDef
}

func (p *Schema) handlePrimaryKeysForCreateTable(tableName schema.TableName, options *schema.TableOptions, tableDef *PostgresTableDef) {
	if options.WithoutPrimaryKey {
		return
	}

	pks := []string{"id"}

	if len(options.PrimaryKeys) > 0 {
		pks = options.PrimaryKeys
	}

	options.PrimaryKeys = pks

	if len(pks) == 1 {
		if val, ok := tableDef.columns.Get(pks[0]); ok {
			val.PrimaryKey = true
			val.NotNull = true
		} else if pks[0] != "id" { // only raise error if the primary key is not "id"
			p.Context.RaiseError(fmt.Errorf("primary key column %s is not defined", pks[0]))
		}
	} else {
		for _, column := range pks {
			if val, ok := tableDef.columns.Get(column); ok {
				val.NotNull = true
			} else {
				p.Context.RaiseError(fmt.Errorf("primary key column %s is not defined", column))
			}
		}
	}

	if len(pks) > 1 {
		tableDef.deferCreationAction = append(tableDef.deferCreationAction, func() {
			p.AddPrimaryKeyConstraint(tableName, pks, schema.PrimaryKeyConstraintOptions{})
		})
	}
}

// PostgresTableDef holds the definition in the table creation.
// create table articles ( <inner_table> )
type PostgresTableDef struct {
	parent              *Schema
	table               schema.TableName
	columns             *orderedmap.OrderedMap[*schema.ColumnOptions]
	innerTable          []string
	deferCreationAction []func()
}

func (p *PostgresTableDef) Columns() []schema.ColumnOptions {
	var columns []schema.ColumnOptions
	for entry := range p.columns.Iterate() {
		columns = append(columns, *entry.Value)
	}
	return columns
}

func (p *PostgresTableDef) AfterTableCreate() []func() {
	return p.deferCreationAction
}

func (p *PostgresTableDef) AddColumn(columnName string, columnType schema.ColumnType, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(columnType, &options)
	p.columns.Set(columnName, &options)
}

func (p *PostgresTableDef) String(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeText, options)
}

func (p *PostgresTableDef) Text(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeText, options)
}

func (p *PostgresTableDef) Integer(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeInteger, options)
}

func (p *PostgresTableDef) BigInt(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeBigInt, options)
}

func (p *PostgresTableDef) Float(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeFloat, options)
}

func (p *PostgresTableDef) Decimal(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeDecimal, options)
}

func (p *PostgresTableDef) Boolean(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeBoolean, options)
}

func (p *PostgresTableDef) DateTime(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeDatetime, options)
}

func (p *PostgresTableDef) Time(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeTime, options)
}

func (p *PostgresTableDef) Date(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeDate, options)
}

func (p *PostgresTableDef) Binary(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeBinary, options)
}

func (p *PostgresTableDef) JSON(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeJSON, options)
}

func (p *PostgresTableDef) UUID(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeUUID, options)
}

func (p *PostgresTableDef) Hstore(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeHstore, options)
}

func (p *PostgresTableDef) Serial(columnName string, opts ...schema.ColumnOptions) {
	options := schema.ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, schema.ColumnTypeSerial, options)
}

// Timestamps adds created_at, updated_at Columns to the table.
func (p *PostgresTableDef) Timestamps() {
	p.AddColumn("created_at", schema.ColumnTypeDatetime, schema.ColumnOptions{NotNull: true, Default: "now()"})
	p.AddColumn("updated_at", schema.ColumnTypeDatetime, schema.ColumnOptions{NotNull: true, Default: "now()"})
}

func (p *PostgresTableDef) Index(columnNames []string, opts ...schema.IndexOptions) {
	options := schema.IndexOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	p.deferCreationAction = append(p.deferCreationAction, func() {
		p.parent.AddIndexConstraint(p.table, columnNames, options)
	})
}

func (p *PostgresTableDef) ForeignKey(toTable schema.TableName, opts ...schema.AddForeignKeyConstraintOptions) {
	options := schema.AddForeignKeyConstraintOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	p.deferCreationAction = append(p.deferCreationAction, func() {
		p.parent.AddForeignKeyConstraint(p.table, toTable, options)
	})
}

// DropTable drops a table from the database.
//
// Example:
//
//	p.DropTable("users", schema.DropTableOptions{})
//
// Generates:
//
//	DROP TABLE "users"
//
// To drop a table if it exists:
//
//	p.DropTable("users", schema.DropTableOptions{IfExists: true})
//
// Generates:
//
//	DROP TABLE IF EXISTS "users"
//
// To make the drop table reversible:
//
//		p.DropTable("users", schema.DropTableOptions{Reversible: &TableOption{
//			schema.TableName: "users",
//			TableDefinition: Innerschema.Tablefunc(t *PostgresTableDef) {
//	         	t.Serial("id")
//				t.FormatRecords("name")
//			}),
//		}})
//
// Generates:
//
//	CREATE TABLE "users" ( "id" serial PRIMARY KEY, "name" text )
func (p *Schema) DropTable(tableName schema.TableName, opts ...schema.DropTableOptions) {
	options := schema.DropTableOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		if options.Reversible != nil {
			p.rollbackMode().CreateTable(tableName, func(t *PostgresTableDef) {}, *options.Reversible)
		} else {
			logger.Warn(events.MessageEvent{
				Message: fmt.Sprintf("unable to reverse dropping table %s", tableName.String()),
			})
		}
		return
	}

	q := `DROP TABLE {if_exists} {table_name}`

	replacer := utils.Replacer{
		"if_exists":  utils.StrFuncPredicate(options.IfExists, "IF EXISTS"),
		"table_name": utils.StrFunc(tableName.String()),
	}

	_, err := p.DB.ExecContext(p.Context.Context, replacer.Replace(q))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while dropping table: %w", err))
		return
	}

	p.Context.AddTableDropped(schema.DropTableOptions{Table: tableName})
}

// RenameTable renames a table in the database.
//
// Example:
//
//	p.RenameTable("users", "people")
//
// Generates:
//
//	ALTER TABLE "users" RENAME TO "people"
func (p *Schema) RenameTable(oldTableName, newTableName schema.TableName) {
	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		p.rollbackMode().RenameTable(newTableName, oldTableName)
		return
	}

	q := `ALTER TABLE {old_table_name} RENAME TO {new_table_name}`

	replacer := utils.Replacer{
		"old_table_name": utils.StrFunc(oldTableName.String()),
		"new_table_name": utils.StrFunc(newTableName.String()),
	}

	_, err := p.DB.ExecContext(p.Context.Context, replacer.Replace(q))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while renaming table: %w", err))
		return
	}

	p.Context.AddTableRenamed(schema.RenameTableOptions{OldTable: oldTableName, NewTable: newTableName})
}
