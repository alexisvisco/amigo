package entrypoint

import (
	"database/sql"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/schema"
)

type DatabaseProvider func(cfg amigoconfig.Config) (*sql.DB, error)

type MainOptions struct {
	CustomAmigo func(a *amigo.Amigo) amigo.Amigo
}

type MainOptFn func(options *MainOptions)

func Main(db DatabaseProvider, migrationsList []schema.Migration, opts ...MainOptFn) {
	database = db
	migrations = migrationsList

	options := &MainOptions{}
	for _, opt := range opts {
		opt(options)
	}

	_ = rootCmd.Execute()
}

func WithCustomAmigo(f func(a *amigo.Amigo) amigo.Amigo) MainOptFn {
	return func(options *MainOptions) {
		options.CustomAmigo = f
	}
}
