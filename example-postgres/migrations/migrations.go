package migrations

import (
	"embed"

	"github.com/alexisvisco/amigo"
)

//go:embed *.sql
var sqlFiles embed.FS

func Migrations(cfg amigo.Configuration) []amigo.Migration {
	return []amigo.Migration{
		amigo.SQLFileToMigration(sqlFiles, "20260116120000_test_multiple_migrations.sql", cfg),
	}
}
