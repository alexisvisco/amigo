package migrations

import "github.com/alexisvisco/amigo"

func Migrations(cfg amigo.Configuration) []amigo.Migration {
	return []amigo.Migration{
		amigo.SQLFileToMigration("20251224200946_create_users.sql", cfg),
		amigo.SQLFileToMigration("20251224200947_add_posts.sql", cfg),
		amigo.SQLFileToMigration("20251224200949_add_comments.sql", cfg),
		amigo.SQLFileToMigration("20251224202202_test.sql", cfg),
	}
}
