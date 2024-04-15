package schema

type PostgresTableDef struct {
	parent             *Postgres
	columns            []ColumnOptions
	afterCreatingTable []func()
}

func (p *PostgresTableDef) AddColumn(columnName string, columnType ColumnType, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(columnType, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) String(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeString, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) Text(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeText, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) Integer(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeInteger, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) BigInt(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeBigInt, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) Float(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeFloat, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) Decimal(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeDecimal, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) Boolean(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeBoolean, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) DateTime(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeDatetime, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) Time(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeTime, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) Date(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeDate, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) Binary(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeBinary, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) JSON(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeJSON, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) UUID(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeUUID, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) Hstore(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeHstore, &options)
	p.columns = append(p.columns, options)
}

func (p *PostgresTableDef) Serial(columnName string, options ColumnOptions) {
	options.ColumnName = columnName
	options.ColumnType = p.parent.toType(ColumnTypeSerial, &options)
	p.columns = append(p.columns, options)
}
