package entrypoint

import (
	"database/sql"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoconfig"
	"github.com/alexisvisco/amigo/pkg/schema"
)

type Provider func(cfg amigoconfig.Config) (*sql.DB, []schema.Migration, error)

func Main(resourceProvider Provider, opts ...amigo.OptionFn) {
	provider = resourceProvider
	amigoOptions = opts

	_ = rootCmd.Execute()
}
