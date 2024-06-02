package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
)

// IndexExist checks if the specified index exists for the given table
func (p *Schema) IndexExist(tableName schema.TableName, indexName string) bool {
	query := `SELECT 1 FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND name = ?`
	var exists int
	err := p.TX.QueryRowContext(p.Context.Context, query, tableName.Name(), indexName).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		p.Context.RaiseError(fmt.Errorf("error while checking if index exists: %w", err))
		return false
	}
	return exists == 1
}
