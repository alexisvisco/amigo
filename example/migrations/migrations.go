package migrations

import (
	"embed"

	"github.com/alexisvisco/amigo"
)

//go:embed *.sql
var sqlFiles embed.FS

func Migrations(cfg amigo.Configuration) []amigo.Migration {
	return []amigo.Migration{
		amigo.SQLFileToMigration(sqlFiles, "20251224200647_create_users.sql", cfg),
		amigo.SQLFileToMigration(sqlFiles, "20251224200648_create_posts.sql", cfg),
		amigo.SQLFileToMigration(sqlFiles, "20251224200650_create_comments.sql", cfg),
		&Migration20251225101501AddUserEmailIndex{},
	}
}
