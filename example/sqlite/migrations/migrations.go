// Package migrations
// /!\ File is auto-generated DO NOT EDIT.
package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema"
)

var Migrations = []schema.Migration{
	&Migration20240602080728CreateTableSchemaVersion{},
	&Migration20240602081304AddIndex{},
	&Migration20240602081806DropIndex{},
}
