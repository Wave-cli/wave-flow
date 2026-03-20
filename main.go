// wave-flow is a wave CLI plugin for development workflow automation.
//
// It reads command definitions from the [flow] section of a Wavefile
// and executes them with optional environment variables, success/failure
// callbacks, and watch mode.
//
// Usage:
//
//	wave flow <command> [args...]
//	wave flow --list
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/wave-cli/wave-core/pkg/sdk"
	"github.com/wave-cli/wave-flow/internal/flow"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run is the testable core of the plugin. It reads config from r,
// processes args, writes output to stdout/stderr, and returns an exit code.
func run(args []string, r io.Reader, stdout, stderr io.Writer) int {
	// Handle -h / --help flag first (before reading config)
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		printHelp(stdout)
		return 0
	}

	// Read config via SDK
	cfg, err := sdk.ReadConfigFrom(r)
	if err != nil {
		sdk.FormatWaveError(stderr, "flow-config-error", "failed to read config", err.Error())
		return 1
	}
	config := cfg.Raw()

	// Handle --list flag
	if len(args) > 0 && (args[0] == "--list" || args[0] == "-l") {
		cmds := flow.ListCommands(config)
		if len(cmds) == 0 {
			fmt.Fprintln(stdout, "No flow commands defined. Add commands to the [flow] section of your Wavefile.")
			return 0
		}
		fmt.Fprintln(stdout, "Available flow commands:")
		for _, name := range cmds {
			fmt.Fprintf(stdout, "  %s\n", name)
		}
		return 0
	}

	// Require a command name
	if len(args) == 0 {
		fmt.Fprintln(stdout, "No command specified. Run 'wave flow --help' for usage.")
		return 0
	}

	cmdName := args[0]

	// Resolve and execute the command
	cmd, err := flow.ResolveCommand(config, cmdName)
	if err != nil {
		sdk.FormatWaveError(stderr, "flow-resolve-error", err.Error(), "Check your Wavefile [flow] section.")
		return 1
	}

	return flow.RunCommand(cmd, stdout, stderr)
}

// printHelp outputs usage information for wave flow.
func printHelp(w io.Writer) {
	fmt.Fprintln(w, `wave flow - development workflow automation

Usage:
  wave flow <command> [args...]
  wave flow --list

Flags:
  -l, --list    List all available flow commands
  -h, --help    Show this help message

Examples:
  wave flow build       Run the 'build' command
  wave flow dev         Run the 'dev' command
  wave flow --list      List all flow commands

Commands are defined in the [flow] section of your Wavefile:

  [flow.build]
  cmd = "go build ./..."

  [flow.dev]
  cmd = "go run ."
  env = { PORT = "3000" }`)
}
