// Package migrations
// /!\ File is auto-generated DO NOT EDIT.
package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema"
)

var Migrations = []schema.Migration{
	&Migration20240524090427CreateTableSchemaVersion{},
	&Migration20240524090434CreateUserTable{},
}
