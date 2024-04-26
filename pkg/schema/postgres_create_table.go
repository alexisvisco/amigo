package schema

import (
	"fmt"
	"strings"
)

type PostgresTableDef struct {
	parent              *Postgres
	table               TableName
	columns             map[string]*ColumnOptions
	innerTable          []string
	deferCreationAction []func()
}

func (p *Postgres) CreateTable(tableName TableName, f func(*PostgresTableDef), opts ...CreateTableOptions) {
	options := CreateTableOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	options.Table = tableName

	if p.Context.migrationType == MigrationTypeDown {
		p.DropTable(tableName, DropTableOptions{IfExists: true})
		return
	}

	var td *PostgresTableDef
	if options.PostgresTableDefinition != nil {
		td = options.PostgresTableDefinition
	} else {
		td = p.BuildInnerTable(tableName, f, options)
	}

	q := `CREATE TABLE {if_not_exists} {table_name} (
		{inner_table}
    ) {table_options}`

	replacer := replacer{
		"if_not_exists": func() string {
			if options.IfNotExists {
				return "IF NOT EXISTS"
			}
			return ""
		},

		"table_name": strfunc(tableName.String()),

		"inner_table": func() string {
			return strings.Join(td.innerTable, ",\n\t\t")
		},

		"table_options": func() string {
			if options.Option != "" {
				return options.Option
			}
			return ""
		},
	}

	_, err := p.db.ExecContext(p.Context.Context, replacer.replace(q))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while creating table: %w", err))
		return
	}

	for _, afterCreate := range td.deferCreationAction {
		afterCreate()
	}
}

func (p *Postgres) BuildInnerTable(tableName TableName, f func(*PostgresTableDef), options CreateTableOptions) *PostgresTableDef {
	tableDef := &PostgresTableDef{
		parent:  p,
		table:   tableName,
		columns: map[string]*ColumnOptions{},
	}

	f(tableDef)

	p.handlePrimaryKeysForCreateTable(tableName, options, tableDef)

	var innerTable []string

	for _, options := range tableDef.columns {
		if options.PrimaryKey {
			options.NotNull = true
			options.Constraints = append(options.Constraints, PrimaryKeyConstraintOptions{})
		}
		innerTable = append(innerTable, p.column(*options))
	}

	tableDef.innerTable = innerTable

	return tableDef
}

func (p *Postgres) handlePrimaryKeysForCreateTable(tableName TableName, options CreateTableOptions, tableDef *PostgresTableDef) {
	pks := []string{"id"}

	if len(options.PrimaryKeys) > 0 {
		pks = options.PrimaryKeys
	}

	if len(pks) == 1 {
		if val, ok := tableDef.columns[pks[0]]; ok {
			val.PrimaryKey = true
			val.NotNull = true
		} else {
			p.Context.RaiseError(fmt.Errorf("primary key column %s is not defined", pks[0]))
		}
	} else {
		for _, column := range pks {
			if val, ok := tableDef.columns[column]; ok {
				val.NotNull = true
			} else {
				p.Context.RaiseError(fmt.Errorf("primary key column %s is not defined", column))
			}
		}
	}

	if len(pks) > 1 {
		tableDef.deferCreationAction = append(tableDef.deferCreationAction, func() {
			p.AddPrimaryKeyConstraint(tableName, pks, PrimaryKeyConstraintOptions{})
		})
	}
}

func (p *PostgresTableDef) AddColumn(columnName string, columnType ColumnType, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(columnType, &options)
	p.columns[columnName] = &options
}

func (p *PostgresTableDef) String(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeText, options)
}

func (p *PostgresTableDef) Text(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeText, options)
}

func (p *PostgresTableDef) Integer(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeInteger, options)
}

func (p *PostgresTableDef) BigInt(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeBigInt, options)
}

func (p *PostgresTableDef) Float(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeFloat, options)
}

func (p *PostgresTableDef) Decimal(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeDecimal, options)
}

func (p *PostgresTableDef) Boolean(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeBoolean, options)
}

func (p *PostgresTableDef) DateTime(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeDatetime, options)
}

func (p *PostgresTableDef) Time(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeTime, options)
}

func (p *PostgresTableDef) Date(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeDate, options)
}

func (p *PostgresTableDef) Binary(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeBinary, options)
}

func (p *PostgresTableDef) JSON(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeJSON, options)
}

func (p *PostgresTableDef) UUID(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeUUID, options)
}

func (p *PostgresTableDef) Hstore(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeHstore, options)
}

func (p *PostgresTableDef) Serial(columnName string, opts ...ColumnOptions) {
	options := ColumnOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	p.AddColumn(columnName, ColumnTypeSerial, options)
}

// Timestamps adds created_at, updated_at columns to the table.
func (p *PostgresTableDef) Timestamps() {
	p.AddColumn("created_at", ColumnTypeDatetime, ColumnOptions{Default: "now()"})
	p.AddColumn("updated_at", ColumnTypeDatetime, ColumnOptions{Default: "now()"})
}

func (p *PostgresTableDef) Index(columnNames []string, opts ...IndexOptions) {
	options := IndexOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	p.deferCreationAction = append(p.deferCreationAction, func() {
		p.parent.AddIndexConstraint(p.table, columnNames, options)
	})
}

func (p *PostgresTableDef) ForeignKey(toTable TableName, opts ...AddForeignKeyConstraintOptions) {
	options := AddForeignKeyConstraintOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	p.deferCreationAction = append(p.deferCreationAction, func() {
		p.parent.AddForeignKeyConstraint(p.table, toTable, options)
	})
}

// DropTable drops a table from the database.
func (p *Postgres) DropTable(tableName TableName, opts ...DropTableOptions) {
	options := DropTableOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	if p.Context.migrationType == MigrationTypeDown && options.Reversible != nil {
		p.CreateTable(tableName, func(t *PostgresTableDef) {}, CreateTableOptions{})
		return
	}

	q := `DROP TABLE {if_exists} {table_name}`

	replacer := replacer{
		"if_exists": func() string {
			if options.IfExists {
				return "IF EXISTS"
			}
			return ""
		},

		"table_name": strfunc(tableName.String()),
	}

	_, err := p.db.ExecContext(p.Context.Context, replacer.replace(q))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while dropping table: %w", err))
		return
	}
}
