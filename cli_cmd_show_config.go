package amigo

import (
	"flag"
	"fmt"
	"text/tabwriter"
)

// cliShowConfig displays the current configuration
func (c *CLI) cliShowConfig(args []string) int {
	// Show help if requested (before parsing to avoid flag.Parse handling it)
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		c.cliShowConfigHelp()
		return 0
	}

	fs := flag.NewFlagSet("show-config", flag.ContinueOnError)
	fs.SetOutput(c.errorOutput)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	w := tabwriter.NewWriter(c.output, 0, 0, 2, ' ', 0)
	defer w.Flush()

	driverName := "unknown"
	if c.config.Driver != nil {
		driverName = fmt.Sprintf("%s", c.config.Driver.Name())
	}

	fmt.Fprintln(w, "Setting\tValue")
	fmt.Fprintf(w, "Driver\t%+v\n", driverName)
	fmt.Fprintf(w, "DatabaseConnected\t%v\n", c.config.DB != nil)
	fmt.Fprintf(w, "SQLFileUpAnnotation\t%s\n", c.config.SQLFileUpAnnotation)
	fmt.Fprintf(w, "SQLFileDownAnnotation\t%s\n", c.config.SQLFileDownAnnotation)
	fmt.Fprintf(w, "CLI.Directory\t%s\n", c.cliOutput.path(c.directory))
	fmt.Fprintf(w, "CLI.DefaultTransactional\t%v\n", c.defaultTransactional)
	fmt.Fprintf(w, "CLI.MigrationsLoaded\t%d\n", len(c.migrations))

	return 0
}

// cliShowConfigHelp displays help for the show-config command
func (c *CLI) cliShowConfigHelp() {
	help := `Usage: show-config [options]

Display the current migration configuration including directory, annotations,
and loaded migrations count.

Options:
  -h, --help    Show this help message

Examples:
  show-config   Display current configuration
`
	fmt.Fprint(c.output, help)
}
