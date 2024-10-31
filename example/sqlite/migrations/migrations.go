// Package migrations
// /!\ File is auto-generated DO NOT EDIT.
package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/pg"

	"embed"
)

//go:embed *.sql
var sqlMigrationsFS embed.FS

var Migrations = []schema.Migration{
	&Migration20240602080728CreateTableSchemaVersion{},
	&Migration20240602081304AddIndex{},
	&Migration20240602081806DropIndex{},
	schema.NewSQLMigration[*pg.Schema](sqlMigrationsFS, "20240602081806_drop_index.sql", "2024-06-02T10:18:06+02:00",
		"-- migrate:down"),
}
