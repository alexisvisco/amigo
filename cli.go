package amigo

import (
	"fmt"
	"io"
	"os"
)

// CLI represents the command-line interface for migrations
type CLI struct {
	config      Configuration
	runner      *Runner
	migrations  []Migration
	output      io.Writer
	errorOutput io.Writer
}

// CLIConfig holds the configuration for creating a CLI instance
type CLIConfig struct {
	Config     Configuration
	Migrations []Migration
	Output     io.Writer // defaults to os.Stdout
	ErrorOut   io.Writer // defaults to os.Stderr
}

// NewCLI creates a new CLI instance with the given configuration
func NewCLI(cfg CLIConfig) *CLI {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	if cfg.ErrorOut == nil {
		cfg.ErrorOut = os.Stderr
	}

	runner := NewRunner(cfg.Config)

	return &CLI{
		config:      cfg.Config,
		runner:      &runner,
		migrations:  cfg.Migrations,
		output:      cfg.Output,
		errorOutput: cfg.ErrorOut,
	}
}

// Run executes the CLI with the given arguments
// This is the main entry point that should be called from your main function
func (c *CLI) Run(args []string) int {
	if len(args) == 0 {
		c.cliPrintHelp()
		return 0
	}

	cmd := args[0]

	switch cmd {
	case "help", "-h", "--help":
		c.cliPrintHelp()
		return 0
	case "show-config":
		return c.cliShowConfig(args[1:])
	case "generate":
		return c.cliGenerate(args[1:])
	case "up":
		return c.cliUp(args[1:])
	case "down":
		return c.cliDown(args[1:])
	case "status":
		return c.cliStatus(args[1:])
	default:
		fmt.Fprintf(c.errorOutput, "Unknown command: %s\n\n", cmd)
		c.cliPrintHelp()
		return 1
	}
}

// cliPrintHelp displays the help message
func (c *CLI) cliPrintHelp() {
	help := `Usage: [command] [options]

Commands:
  help          Show this help message
  show-config   Display current migration configuration
  generate      Generate a new migration file
  up            Run pending migrations
  down          Revert applied migrations
  status        Show migration status

Options:
  -h, --help    Show help for a command

Run '[command] --help' for more information on a command.
`
	fmt.Fprint(c.output, help)
}
