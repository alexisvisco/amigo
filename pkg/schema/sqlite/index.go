package sqlite

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"strings"
)

// AddIndex adds a new index to the Table. Columns is a list of column names to index.
//
// Creating a simple index:
//
//	p.AddIndex("products", []string{"name"}, IndexOptions{})
//
// Generates:
//
//	CREATE INDEX idx_products_name ON "products" (name)
//
// Creating a unique index:
//
//	p.AddIndex("products", []string{"name"}, IndexOptions{Unique: true})
//
// Generates:
//
//	CREATE UNIQUE INDEX idx_products_name ON "products" (name)
//
// Creating an index with a custom name:
//
//	p.AddIndex("products", []string{"name"}, IndexOptions{IndexNameBuilder: func(Table schema.TableName, Columns []string) string {
//		return "index_products_on_name"
//	}})
//
// Generates:
//
//	CREATE INDEX index_products_on_name ON "products" (name)
//
// Creating an index if it does not exist:
//
//	p.AddIndex("products", []string{"name"}, IndexOptions{IfNotExists: true})
//
// Generates:
//
//	CREATE INDEX IF NOT EXISTS idx_products_name ON "products" (name)
//
// Creating an index with a method:
//
//	p.AddIndex("products", []string{"name"}, IndexOptions{Method: "btree"})
//
// Generates:
//
//	CREATE INDEX idx_products_name ON "products" USING btree (name)
//
// Creating an index with a custom order:
//
//	p.AddIndex("products", []string{"name"}, IndexOptions{Order: "DESC"})
//
// Generates:
//
//	CREATE INDEX idx_products_name ON "products" (name DESC)
//
// Creating an index with custom order per column:
//
//	p.AddIndex("products", []string{"name", "price"}, IndexOptions{OrderPerColumn: map[string]string{"name": "DESC"}})
//
// Generates:
//
//	CREATE INDEX idx_products_name_price ON "products" (name DESC, price)
//
// Creating an index with a predicate:
//
//	p.AddIndex("products", []string{"name"}, IndexOptions{Predicate: "name IS NOT NULL"})
//
// Generates:
//
//	CREATE INDEX idx_products_name ON "products" (name) WHERE name IS NOT NULL
func (p *Schema) AddIndex(table schema.TableName, columns []string, option ...schema.IndexOptions) {
	options := schema.IndexOptions{}
	if len(option) > 0 {
		options = option[0]
	}

	options.Table = table
	options.Columns = columns
	options.IndexName = options.BuildIndexName(options.Table, options.Columns)

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		p.rollbackMode().DropIndex(options.Table, options.Columns, schema.DropIndexOptions{IfExists: true})
		return
	}

	if options.Concurrent {
		logger.Warn(events.MessageEvent{Message: "sqlite does not support concurrent index creation"})
	}

	if options.Method != "" {
		logger.Warn(events.MessageEvent{Message: "sqlite does not support index method (USING)"})
	}

	sql := `CREATE {unique} INDEX {if_not_exists} {index_name} ON {table_name} ({Columns}) {where}`

	replacer := utils.Replacer{
		"unique": func() string {
			if options.Unique {
				return "UNIQUE"
			}
			return ""
		},

		"if_not_exists": func() string {
			if options.IfNotExists {
				return "IF NOT EXISTS"
			}
			return ""
		},

		"index_name": utils.StrFunc(options.IndexName),

		"table_name": utils.StrFunc(options.Table.Name()),

		"Columns": func() string {
			column := "{name} {order}"

			cols := make([]string, len(options.Columns))

			for i, v := range options.Columns {
				replacer := utils.Replacer{
					"name": utils.StrFunc(v),
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

				cols[i] = replacer.Replace(column)
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

	_, err := p.TX.ExecContext(p.Context.Context, replacer.Replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while adding index: %w", err))
		return
	}

	p.Context.AddIndexCreated(options)
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
//
// Dropping an index if it exists:
//
//	p.DropIndex("products", []string{"name"}, DropIndexOptions{IfExists: true})
//
// Generates:
//
//	DROP INDEX IF EXISTS idx_products_name
//
// To reverse the operation, you can use the reversible option:
//
//	p.DropIndex("products", []string{"name"}, DropIndexOptions{
//		Reversible: &schema.IndexOptions{}
//	})
//
// Generates:
//
//	CREATE INDEX idx_products_name ON "products" (name)
func (p *Schema) DropIndex(table schema.TableName, columns []string, opt ...schema.DropIndexOptions) {
	options := schema.DropIndexOptions{}
	if len(opt) > 0 {
		options = opt[0]
	}

	options.Table = table
	options.Columns = columns
	options.IndexName = options.BuildIndexName(options.Table, options.Columns)

	p.BaseDropIndex(options, func(table schema.TableName, columns []string, opts schema.IndexOptions) {
		p.rollbackMode().AddIndex(table, columns, opts)
	})
}
