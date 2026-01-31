package amigo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CLI represents the command-line interface for migrations
type CLI struct {
	config               Configuration
	runner               *Runner
	migrations           []Migration
	output               io.Writer
	errorOutput          io.Writer
	cliOutput            *cliOutput
	directory            string
	defaultTransactional bool
	defaultFileFormat    string
	packageName          string
}

// CLIConfig holds the configuration for creating a CLI instance
type CLIConfig struct {
	// Config is the migration configuration
	Config Configuration

	// Migrations is the list of available migrations
	Migrations []Migration

	// Output is the writer for standard output
	Output io.Writer // defaults to os.Stdout

	// ErrorOut is the writer for error messages
	ErrorOut io.Writer // defaults to os.Stderr

	// Directory is the location of the migrations files
	Directory string

	// DefaultTransactional indicates if new migrations should be run inside a transaction by wrapping them in a Tx helper
	// or putting the tx annotation in SQL files
	DefaultTransactional bool

	// DefaultFileFormat is the default file format for new migrations (sql or go)
	DefaultFileFormat string

	// PackageName is the package name for generated Go files. If not specified, defaults to the directory name.
	PackageName string
}

// NewCLI creates a new CLI instance with the given configuration
func NewCLI(cfg CLIConfig) *CLI {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	if cfg.ErrorOut == nil {
		cfg.ErrorOut = os.Stderr
	}

	// Use the folder name as package name if not specified
	packageName := cfg.PackageName
	if packageName == "" && cfg.Directory != "" {
		packageName = filepath.Base(cfg.Directory)
	}
	if packageName == "" {
		packageName = "migrations" // fallback default
	}

	runner := NewRunner(cfg.Config)

	return &CLI{
		config:               cfg.Config,
		runner:               runner,
		migrations:           cfg.Migrations,
		output:               cfg.Output,
		errorOutput:          cfg.ErrorOut,
		cliOutput:            newCLIOutput(),
		directory:            cfg.Directory,
		defaultTransactional: cfg.DefaultTransactional,
		defaultFileFormat:    cfg.DefaultFileFormat,
		packageName:          packageName,
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
