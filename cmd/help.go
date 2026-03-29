package cmd

import (
	"fmt"
	"io"
)

// PrintHelp outputs usage information for wave flow.
func PrintHelp(w io.Writer) {
	fmt.Fprintln(w, `wave flow - development workflow automation

Usage:
  wave flow <command> [flags]
  wave flow --list
  wave flow --version

Flags:
  -l, --list    List all available flow commands
  -w, --watch   Run command in watch mode (restart on file changes)
  -v, --version Show version information
  -h, --help    Show this help message

Examples:
  wave flow build         Run the 'build' command
  wave flow dev           Run the 'dev' command
  wave flow dev --watch   Run 'dev' in watch mode
  wave flow --list        List all flow commands
  wave flow --version     Show version

Commands are defined in the [flow] section of your Wavefile:

  [flow.build]
  cmd = "go build ./..."

  [flow.dev]
  cmd = "go run ."
  watch = ["*.go", "*.mod"]
  env = { PORT = "3000" }`)
}
