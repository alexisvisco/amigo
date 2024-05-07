package pg

import (
	"github.com/alexisvisco/mig/pkg/schema"
	"github.com/alexisvisco/mig/pkg/utils"
	"testing"
)

func TestPostgresSchema_ConstraintExist(t *testing.T) {
	t.Parallel()

	sc := "tst_pg_constraint_exist"

	query := `CREATE TABLE IF NOT EXISTS {schema}.{table_name} (
			id serial PRIMARY KEY, 
			name text
			constraint {constraint_name} CHECK (name <> '')
    	);`
	replacer := utils.Replacer{
		"schema":          utils.StrFunc(sc),
		"table_name":      utils.StrFunc("test_table"),
		"constraint_name": utils.StrFunc("test_constraint"),
	}

	p, _, sc := baseTest(t, replacer.Replace(query), sc)

	t.Run("must have a constraint", func(t *testing.T) {
		assertConstraintExist(t, p, schema.Table("test_table", sc), "test_constraint")
	})

	t.Run("must not have a constraint", func(t *testing.T) {
		assertConstraintNotExist(t, p, schema.Table("test_table", sc), "test_constraint_not_exists")
	})
}
