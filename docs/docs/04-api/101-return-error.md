# Return error

As you can see in a migration file the functions Up, Down or Change cannot return an error. 
If you want to raise an error you can use the `RaiseError` function from the context.

```go
package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"...../repositories/userrepo"
	"time"
)

type Migration20240517135429Droptable struct{}

func (m Migration20240517135429Droptable) Change(s *pg.Schema) {
    s.CreateTable("test", func(def *pg.PostgresTableDef) {
        def.String("name")
        def.JSON("data")
    })
	
    _, err := userrepo.New(s.DB).GetUser(5)
    if err != nil {
		s.Context.RaiseError(fmt.Errorf("error: %w", err))
    }
}

func (m Migration20240517135429Droptable) Name() string {
	return "droptable"
}

func (m Migration20240517135429Droptable) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-05-17T15:54:29+02:00")
	return t
}

```

In this example, if the `GetUser` function returns an error, the migration will fail and the error will be displayed in the logs.


The only way to not crash the migration when a RaiseError is called is to use the `--continue-on-error` flag.

And the only way to crash when this flag is used is to use a `schema.ForceStopError` error.

```go
s.Context.RaiseError(schema.NewForceStopError(errors.New("force stop")))
```

This will crash the migration EVEN if the `--continue-on-error` flag is used.
