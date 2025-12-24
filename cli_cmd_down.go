package amigo

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
)

// cliDown reverts applied migrations
func (c *CLI) cliDown(args []string) int {
	// Show help if requested (before parsing to avoid flag.Parse handling it)
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		c.cliDownHelp()
		return 0
	}

	fs := flag.NewFlagSet("down", flag.ContinueOnError)
	fs.SetOutput(c.errorOutput)

	var steps int
	var autoConfirm bool
	fs.IntVar(&steps, "steps", 1, "Number of migrations to revert (default: 1)")
	fs.BoolVar(&autoConfirm, "yes", false, "Skip confirmation prompt")
	fs.BoolVar(&autoConfirm, "y", false, "Skip confirmation prompt (shorthand)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	ctx := context.Background()

	// Get migration statuses to show what will be reverted
	statuses, err := c.runner.GetMigrationsStatuses(ctx, c.migrations)
	if err != nil {
		fmt.Fprintf(c.errorOutput, "Error: failed to get migration statuses: %v\n", err)
		return 1
	}

	var appliedMigrations []MigrationStatus
	for _, status := range statuses {
		if status.Applied {
			appliedMigrations = append(appliedMigrations, status)
		}
	}

	// Sort by date descending (newest first)
	slices.SortFunc(appliedMigrations, func(a, b MigrationStatus) int {
		if a.Migration.Date > b.Migration.Date {
			return -1
		} else if a.Migration.Date < b.Migration.Date {
			return 1
		}
		return 0
	})

	if len(appliedMigrations) == 0 {
		fmt.Fprintln(c.output, "No applied migrations to revert")
		return 0
	}

	// Determine how many migrations will be reverted
	migrationsToRevert := appliedMigrations
	if steps > 0 && steps < len(appliedMigrations) {
		migrationsToRevert = appliedMigrations[:steps]
	}

	// Display migrations to revert
	fmt.Fprintf(c.output, "The following %d migration(s) will be reverted:\n\n", len(migrationsToRevert))

	w := tabwriter.NewWriter(c.output, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Date\tName")
	for _, m := range migrationsToRevert {
		fmt.Fprintf(w, "%d\t%s\n", m.Migration.Date, m.Migration.Name)
	}
	w.Flush()

	fmt.Fprintln(c.output, "")

	// Prompt for confirmation unless --yes flag is set
	if !autoConfirm {
		fmt.Fprint(c.output, "Do you want to continue? (yes/no): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(c.errorOutput, "Error: failed to read input: %v\n", err)
			return 1
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			fmt.Fprintln(c.output, "Migration cancelled")
			return 0
		}
	}

	// Build options
	var opts []RunnerDownOptsFunc
	if steps >= 0 {
		opts = append(opts, RunnerDownOptionSteps(steps))
	}

	// Run migrations using iterator to show progress
	fmt.Fprintln(c.output, "")
	migrationCount := 0
	for result := range c.runner.DownIterator(ctx, c.migrations, opts...) {
		if result.Error != nil {
			fmt.Fprintf(c.errorOutput, "Error: %v\n", result.Error)
			return 1
		}

		migrationCount++
		fmt.Fprintf(c.output, "== %s: reverting (%.2fs)\n", result.Migration.Name(), result.Duration.Seconds())
	}

	fmt.Fprintln(c.output, "")
	fmt.Fprintf(c.output, "Successfully reverted %d migration(s)\n", migrationCount)
	return 0
}

// cliDownHelp displays help for the down command
func (c *CLI) cliDownHelp() {
	help := `Usage: down [options]

Revert applied migrations.

Options:
  --steps int    Number of migrations to revert (default: 1)
  -y, --yes      Skip confirmation prompt
  -h, --help     Show this help message

Examples:
  down              Revert the last applied migration
  down --steps=2    Revert the last 2 applied migrations
  down --steps=-1   Revert all applied migrations
  down --yes        Revert without confirmation
`
	fmt.Fprint(c.output, help)
}
