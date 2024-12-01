package entrypoint

import (
	"database/sql"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/schema"
)

type Provider func(cfg amigoconfig.Config) (*sql.DB, []schema.Migration, error)

type MainOptions struct {
	CustomAmigo func(a *amigo.Amigo) amigo.Amigo
}

type MainOptFn func(options *MainOptions)

func Main(resourceProvider Provider, opts ...MainOptFn) {
	provider = resourceProvider

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
