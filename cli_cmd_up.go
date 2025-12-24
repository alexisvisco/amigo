package amigo

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// cliUp runs pending migrations
func (c *CLI) cliUp(args []string) int {
	// Show help if requested (before parsing to avoid flag.Parse handling it)
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		c.cliUpHelp()
		return 0
	}

	fs := flag.NewFlagSet("up", flag.ContinueOnError)
	fs.SetOutput(c.errorOutput)

	var steps int
	var autoConfirm bool
	fs.IntVar(&steps, "steps", -1, "Number of migrations to run (default: all)")
	fs.BoolVar(&autoConfirm, "yes", false, "Skip confirmation prompt")
	fs.BoolVar(&autoConfirm, "y", false, "Skip confirmation prompt (shorthand)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	ctx := context.Background()

	// Get migration statuses to show what will be applied
	statuses, err := c.runner.GetMigrationsStatuses(ctx, c.migrations)
	if err != nil {
		fmt.Fprintf(c.errorOutput, "%s\n", c.cliOutput.error(fmt.Sprintf("Error: failed to get migration statuses: %v", err)))
		return 1
	}

	// Filter pending migrations
	var pendingMigrations []MigrationStatus
	for _, status := range statuses {
		if !status.Applied {
			pendingMigrations = append(pendingMigrations, status)
		}
	}

	if len(pendingMigrations) == 0 {
		fmt.Fprintln(c.output, "No pending migrations to apply")
		return 0
	}

	// Determine how many migrations will be applied
	migrationsToApply := pendingMigrations
	if steps > 0 && steps < len(pendingMigrations) {
		migrationsToApply = pendingMigrations[:steps]
	}

	// Display migrations to apply
	fmt.Fprintf(c.output, "The following %d migration(s) will be applied:\n\n", len(migrationsToApply))

	w := tabwriter.NewWriter(c.output, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Date\tName")
	for _, m := range migrationsToApply {
		fmt.Fprintf(w, "%s\t%s\n", c.cliOutput.date(m.Migration.Date), m.Migration.Name)
	}
	w.Flush()

	fmt.Fprintln(c.output, "")

	// Prompt for confirmation unless --yes flag is set
	if !autoConfirm {
		fmt.Fprint(c.output, "Do you want to continue? (yes/no): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(c.errorOutput, "%s\n", c.cliOutput.error(fmt.Sprintf("Error: failed to read input: %v", err)))
			return 1
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			fmt.Fprintln(c.output, "Migration cancelled")
			return 0
		}
	}

	// Build options
	var opts []RunnerUpOptsFunc
	if steps >= 0 {
		opts = append(opts, RunnerUpOptionSteps(steps))
	}

	// Run migrations using iterator to show progress
	fmt.Fprintln(c.output, "")
	migrationCount := 0
	for result := range c.runner.UpIterator(ctx, c.migrations, opts...) {
		if result.Error != nil {
			fmt.Fprintf(c.errorOutput, "%s\n", c.cliOutput.error(fmt.Sprintf("Error: %v", result.Error)))
			return 1
		}

		migrationCount++
		fmt.Fprintf(c.output, "== %s: migrating (%s)\n\n", result.Migration.Name(), c.cliOutput.duration(result.Duration))
	}

	fmt.Fprintln(c.output, "")
	fmt.Fprintf(c.output, "Successfully applied %d migration(s)\n", migrationCount)
	return 0
}

// cliUpHelp displays help for the up command
func (c *CLI) cliUpHelp() {
	help := `Usage: up [options]

Run pending migrations.

Options:
  --steps int    Number of migrations to run (default: all pending migrations)
  -y, --yes      Skip confirmation prompt
  -h, --help     Show this help message

Examples:
  up              Run all pending migrations
  up --steps=1    Run only the next pending migration
  up --steps=3    Run the next 3 pending migrations
  up --yes        Run all pending migrations without confirmation
`
	fmt.Fprint(c.output, help)
}
