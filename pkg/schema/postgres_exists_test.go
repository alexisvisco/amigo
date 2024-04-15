package schema

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostgresSchema_ConstraintExist(t *testing.T) {
	t.Parallel()

	schema := "tst_pg_constraint_exist"
	conn, _, mig, schema := initSchema(t, schema)

	query := `CREATE TABLE IF NOT EXISTS {schema}.{table_name} (
			id serial PRIMARY KEY, 
			name text
			constraint {constraint_name} CHECK (name <> '')
    	);`
	replacer := replacer{
		"schema":          strfunc(schema),
		"table_name":      strfunc("test_table"),
		"constraint_name": strfunc("test_constraint"),
	}

	_, err := conn.Exec(replacer.replace(query))
	require.NoError(t, err)

	t.Run("must have a constraint", func(t *testing.T) {
		assertConstraintExist(t, mig.newSchema(), Table("test_table", schema), "test_constraint")
	})

	t.Run("must not have a constraint", func(t *testing.T) {
		assertConstraintNotExist(t, mig.newSchema(), Table("test_table", schema), "test_constraint_not_exists")
	})
}
