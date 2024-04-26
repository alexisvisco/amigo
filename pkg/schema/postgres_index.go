package schema

import (
	"fmt"
	"strings"
)

// AddIndexConstraint adds a new index to the Table. Columns is a list of column names to index.
//
// Creating a simple index:
//
//	p.AddIndexConstraint("products", []string{"name"}, IndexOptions{})
//
// Generates:
//
//	CREATE INDEX idx_products_name ON "products" (name)
//
// Creating a unique index:
//
//	p.AddIndexConstraint("products", []string{"name"}, IndexOptions{Unique: true})
//
// Generates:
//
//	CREATE UNIQUE INDEX idx_products_name ON "products" (name)
//
// Creating an index with a custom name:
//
//	p.AddIndexConstraint("products", []string{"name"}, IndexOptions{IndexNameBuilder: func(Table TableName, Columns []string) string {
//		return "index_products_on_name"
//	}})
//
// Generates:
//
//	CREATE INDEX index_products_on_name ON "products" (name)
//
// Creating an index if it does not exist:
//
//	p.AddIndexConstraint("products", []string{"name"}, IndexOptions{IfNotExists: true})
//
// Generates:
//
//	CREATE INDEX IF NOT EXISTS idx_products_name ON "products" (name)
//
// Creating an index with a method:
//
//	p.AddIndexConstraint("products", []string{"name"}, IndexOptions{Method: "btree"})
//
// Generates:
//
//	CREATE INDEX idx_products_name ON "products" USING btree (name)
//
// Creating an index concurrently:
//
//	p.AddIndexConstraint("products", []string{"name"}, IndexOptions{Concurrent: true})
//
// Generates:
//
//	CREATE INDEX CONCURRENTLY idx_products_name ON "products" (name)
//
// Creating an index with a custom order:
//
//	p.AddIndexConstraint("products", []string{"name"}, IndexOptions{Order: "DESC"})
//
// Generates:
//
//	CREATE INDEX idx_products_name ON "products" (name DESC)
//
// Creating an index with custom order per column:
//
//	p.AddIndexConstraint("products", []string{"name", "price"}, IndexOptions{OrderPerColumn: map[string]string{"name": "DESC"}})
//
// Generates:
//
//	CREATE INDEX idx_products_name_price ON "products" (name DESC, price)
//
// Creating an index with a predicate:
//
//	p.AddIndexConstraint("products", []string{"name"}, IndexOptions{Predicate: "name IS NOT NULL"})
//
// Generates:
//
//	CREATE INDEX idx_products_name ON "products" (name) WHERE name IS NOT NULL
func (p *Postgres) AddIndexConstraint(table TableName, columns []string, options IndexOptions) {
	options.Table = table
	options.Columns = columns
	options.IndexName = options.BuildIndexName(options.Table, options.Columns)

	if p.Context.migrationType == MigrationTypeDown {
		p.DropIndex(options.Table, options.Columns, DropIndexOptions{IfExists: true})
		return
	}

	sql := `CREATE {unique} INDEX {concurrently} {if_not_exists} {index_name} ON {table_name} {using} ({Columns}) {where}`

	replacer := replacer{
		"unique": func() string {
			if options.Unique {
				return "UNIQUE"
			}
			return ""
		},

		"concurrently": func() string {
			if options.Concurrent {
				return "CONCURRENTLY"
			}
			return ""
		},

		"if_not_exists": func() string {
			if options.IfNotExists {
				return "IF NOT EXISTS"
			}
			return ""
		},

		"index_name": strfunc(options.IndexName),

		"table_name": strfunc(options.Table.String()),

		"using": func() string {
			if options.Method != "" {
				return fmt.Sprintf("USING %s", options.Method)
			}
			return ""
		},

		"Columns": func() string {
			column := "{name} {order}"

			cols := make([]string, len(options.Columns))

			for i, v := range options.Columns {
				replacer := replacer{
					"name": strfunc(v),
					"order": func() string {
						if order, ok := options.OrderPerColumn[v]; ok {
							return order
						}

						if options.Order != "" {
							return options.Order
						}

						return ""
					},
				}

				cols[i] = replacer.replace(column)
			}

			return strings.Join(cols, ", ")
		},

		"where": func() string {
			if options.Predicate != "" {
				return fmt.Sprintf("WHERE %s", options.Predicate)
			}
			return ""
		},
	}

	_, err := p.db.ExecContext(p.Context.Context, replacer.replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding index: %w", err))
		return
	}

	p.Context.addIndexCreated(options)
}

// DropIndex drops an index from the database.
//
// Example:
//
//	p.DropIndex("products", []string{"name"}, DropIndexOptions{})
//
// Generates:
//
//	DROP INDEX idx_products_name
func (p *Postgres) DropIndex(table TableName, columns []string, options DropIndexOptions) {
	options.Table = table
	options.Columns = columns
	options.IndexName = options.BuildIndexName(options.Table, options.Columns)

	if p.Context.migrationType == MigrationTypeDown && options.Reversible != nil {
		p.AddIndexConstraint(table, columns, *options.Reversible)
		return
	}

	sql := `DROP INDEX {if_exists} {index_name}`
	replacer := replacer{
		"if_exists": func() string {
			if options.IfExists {
				return "IF EXISTS"
			}
			return ""
		},

		"index_name": strfunc(options.IndexName),
	}

	_, err := p.db.ExecContext(p.Context.Context, replacer.replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while dropping index: %w", err))
		return
	}

	p.Context.addIndexDropped(options)
}
