// Package cmd implements the wave-flow CLI commands.
package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/wave-cli/wave-core/pkg/sdk"
	we "github.com/wave-cli/wave-core/pkg/sdk/error"
	"github.com/wave-cli/wave-flow/internal/flow"
)

// Run is the main entry point for the flow plugin.
// It reads config from r, processes args, writes output to stdout/stderr,
// and returns an exit code.
func Run(args []string, r io.Reader, stdout, stderr io.Writer) int {
	// Handle -h / --help flag first (before reading config)
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		PrintHelp(stdout)
		return 0
	}

	// Handle --version flag
	if len(args) > 0 && (args[0] == "--version" || args[0] == "-v") {
		PrintVersion(stdout)
		return 0
	}

	// Read config via SDK
	cfg, err := sdk.ReadConfigFrom(r)
	if err != nil {
		we.Format(stderr, "flow-config-error", "failed to read config", err.Error())
		return 1
	}
	config := cfg.Raw()

	// Handle --list flag
	if len(args) > 0 && (args[0] == "--list" || args[0] == "-l") {
		return ListCommands(config, stdout)
	}

	// Require a command name
	if len(args) == 0 {
		fmt.Fprintf(stdout, "wave-flow %s\n", GetVersion())
		fmt.Fprintln(stdout, "Run 'wave flow --help' for usage.")
		return 0
	}

	// Parse command and flags
	cmdName := args[0]
	watchMode := false

	// Check for --watch flag
	for i, arg := range args {
		if arg == "--watch" || arg == "-w" {
			watchMode = true
			// Remove the flag from args
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	// Resolve and execute the command
	cmd, err := flow.ResolveCommand(config, cmdName)
	if err != nil {
		// Show clean error message without debug codes
		available := flow.ListCommands(config)
		fmt.Fprintf(stderr, "command not found: %s\n", cmdName)
		fmt.Fprintf(stderr, "Available: %s\n", strings.Join(available, ", "))
		fmt.Fprintln(stderr, "Run 'wave flow --list' to see available commands.")
		return 1
	}

	// Run with watch mode if enabled or if command has watch patterns
	if watchMode || len(cmd.Watch) > 0 {
		if len(cmd.Watch) == 0 {
			fmt.Fprintln(stderr, "Error: --watch flag requires watch patterns in command config")
			fmt.Fprintln(stderr, "Add watch = [\"*.go\"] to your command in Wavefile")
			return 1
		}
		return flow.RunWithWatch(cmd, stdout, stderr)
	}

	return flow.RunCommand(cmd, stdout, stderr)
}
