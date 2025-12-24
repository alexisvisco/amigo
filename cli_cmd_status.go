package amigo

import (
	"context"
	"flag"
	"fmt"
	"text/tabwriter"
)

// cliStatus displays the status of all migrations
func (c *CLI) cliStatus(args []string) int {
	// Show help if requested (before parsing to avoid flag.Parse handling it)
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		c.cliStatusHelp()
		return 0
	}

	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(c.errorOutput)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	ctx := context.Background()

	// Get migration statuses
	statuses, err := c.runner.GetMigrationsStatuses(ctx, c.migrations)
	if err != nil {
		fmt.Fprintf(c.errorOutput, "Error: failed to get migration statuses: %v\n", err)
		return 1
	}

	if len(statuses) == 0 {
		fmt.Fprintln(c.output, "No migrations found")
		return 0
	}

	// Count applied and pending
	appliedCount := 0
	pendingCount := 0
	for _, status := range statuses {
		if status.Applied {
			appliedCount++
		} else {
			pendingCount++
		}
	}

	// Display summary
	fmt.Fprintf(c.output, "Migration Status: %d applied, %d pending\n\n", appliedCount, pendingCount)

	// Display migrations table
	w := tabwriter.NewWriter(c.output, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Status\tDate\tName\tApplied At")

	for _, status := range statuses {
		statusStr := "pending"
		appliedAt := ""

		if status.Applied {
			statusStr = "applied"
			appliedAt = status.Migration.AppliedAt.Format("2006-01-02 15:04:05")
		}

		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
			statusStr,
			status.Migration.Date,
			status.Migration.Name,
			appliedAt,
		)
	}

	w.Flush()

	return 0
}

// cliStatusHelp displays help for the status command
func (c *CLI) cliStatusHelp() {
	help := `Usage: status [options]

Show the status of all migrations.

Options:
  -h, --help    Show this help message

Examples:
  status        Display migration status
`
	fmt.Fprint(c.output, help)
}
