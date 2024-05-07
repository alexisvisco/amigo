package pg

import (
	"context"
	"github.com/alexisvisco/mig/pkg/schema"
	"github.com/stretchr/testify/require"
	"testing"
)

func assertTableExist(t *testing.T, p *Schema, table schema.TableName) {
	var exists bool
	err := p.db.QueryRowContext(context.Background(), `SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = $1
		AND table_name = $2
	);`, table.Schema(), table.Name()).Scan(&exists)

	require.NoError(t, err)
	require.True(t, exists)
}

func assertTableNotExist(t *testing.T, p *Schema, table schema.TableName) {
	var exists bool
	err := p.db.QueryRowContext(context.Background(), `SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = $1
		AND table_name = $2
	);`, table.Schema(), table.Name()).Scan(&exists)

	require.NoError(t, err)
	require.False(t, exists)
}
