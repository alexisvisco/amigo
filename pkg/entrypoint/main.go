package entrypoint

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/alexisvisco/amigo/pkg/amigo"
	"github.com/alexisvisco/amigo/pkg/amigoctx"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/colors"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
)

func Main(db *sql.DB, arg amigo.MainArg, migrations []schema.Migration, ctx *amigoctx.Context) {
	am := amigo.NewAmigo(ctx)
	am.SetupSlog(os.Stdout, nil)

	switch arg {
	case amigo.MainArgMigrate, amigo.MainArgRollback:
		dir := types.MigrationDirectionUp
		if arg == amigo.MainArgRollback {
			dir = types.MigrationDirectionDown
		}
		err := am.RunMigrations(amigo.RunMigrationParams{
			DB:         db,
			Direction:  dir,
			Migrations: migrations,
			LogOutput:  os.Stdout,
		})

		if err != nil {
			logger.Error(events.MessageEvent{Message: err.Error()})
			os.Exit(1)
		}
	case amigo.MainArgSkipMigration:
		err := am.SkipMigrationFile(db)
		if err != nil {
			logger.Error(events.MessageEvent{Message: err.Error()})
			os.Exit(1)
		}
	case amigo.MainArgStatus:
		versions, err := am.GetStatus(db)
		if err != nil {
			logger.Error(events.MessageEvent{Message: err.Error()})
			os.Exit(1)
		}

		hasVersion := func(version string) bool {
			for _, v := range versions {
				if v == version {
					return true
				}
			}
			return false
		}

		// show status of 10 last migrations
		b := &strings.Builder{}
		tw := tabwriter.NewWriter(b, 2, 0, 1, ' ', 0)

		defaultMigrations := sliceArrayOrDefault(migrations, 10)

		for i, m := range defaultMigrations {

			key := fmt.Sprintf("(%s) %s", m.Date().UTC().Format(utils.FormatTime), m.Name())
			value := colors.Red("not applied")
			if hasVersion(m.Date().UTC().Format(utils.FormatTime)) {
				value = colors.Green("applied")
			}

			fmt.Fprintf(tw, "%s\t\t%s", key, value)
			if i != len(defaultMigrations)-1 {
				fmt.Fprintln(tw)
			}
		}

		tw.Flush()
		logger.Info(events.MessageEvent{Message: b.String()})
	}
}

func sliceArrayOrDefault[T any](array []T, x int) []T {
	defaultMigrations := array
	if len(array) >= x {
		defaultMigrations = array[len(array)-x:]
	}
	return defaultMigrations
}

func AmigoContextFromFlags() (*amigoctx.Context, amigo.MainArg) {
	jsonFlag := flag.String("json", "", "all amigo context in json | bas64")

	flag.Parse()

	if flag.NArg() == 0 {
		logger.Error(events.MessageEvent{Message: "missing argument"})
		os.Exit(1)
	}

	arg := amigo.MainArg(flag.Arg(0))
	if err := arg.Validate(); err != nil {
		logger.Error(events.MessageEvent{Message: err.Error()})
		os.Exit(1)
	}

	a := amigoctx.NewContext()
	if *jsonFlag != "" {
		b64decoded, err := base64.StdEncoding.DecodeString(*jsonFlag)
		if err != nil {
			logger.Error(events.MessageEvent{Message: fmt.Sprintf("unable to unmarshal amigo context b64: %s",
				err.Error())})
			os.Exit(1)
		}
		err = json.Unmarshal(b64decoded, a)
		if err != nil {
			logger.Error(events.MessageEvent{Message: fmt.Sprintf("unable to unmarshal amigo context json: %s",
				err.Error())})
			os.Exit(1)
		}
	}

	return a, arg
}
