package base

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
)

// BaseDropIndex drops an index from the table. Generic method to drop an index across all databases.
func (p *Schema) BaseDropIndex(
	options schema.DropIndexOptions,
	onRollback func(table schema.TableName, columns []string, opts schema.IndexOptions),
) {

	if p.Context.MigrationDirection == types.MigrationDirectionDown {
		if options.Reversible != nil {
			onRollback(options.Table, options.Columns, *options.Reversible)
		} else {
			logger.Warn(events.MessageEvent{
				Message: fmt.Sprintf("unable re-creating index %s", options.IndexName),
			})
		}
		return
	}

	sql := `DROP INDEX {if_exists} {index_name}`
	replacer := utils.Replacer{
		"if_exists": func() string {
			if options.IfExists {
				return "IF EXISTS"
			}
			return ""
		},

		"index_name": func() string {
			if options.Table.HasSchema() {
				return fmt.Sprintf(`%s.%s`, options.Table.Schema(), options.IndexName)
			}

			return options.IndexName
		},
	}

	_, err := p.TX.ExecContext(p.Context.Context, replacer.Replace(sql))
	if err != nil {
		p.Context.RaiseError(fmt.Errorf("error while dropping index: %w", err))
		return
	}

	p.Context.AddIndexDropped(options)
}
