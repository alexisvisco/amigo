// Package migrations
// /!\ File is auto-generated DO NOT EDIT.
package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema"
)

var Migrations = []schema.Migration{
	&Migration20240517080505SchemaVersion{},
	&Migration20240518071740CreateUser{},
	&Migration20240518071842AddIndexUserEmail{},
	&Migration20240518071938CustomSeed{},
	&Migration20240527192300Enum{},
}
