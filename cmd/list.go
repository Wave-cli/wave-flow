package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/wave-cli/wave-flow/internal/flow"
)

// ListCommands outputs all available flow commands.
// Returns 0 on success.
func ListCommands(config map[string]any, w io.Writer) int {
	cmds := flow.ListCommands(config)
	if len(cmds) == 0 {
		fmt.Fprintln(w, "No flow commands defined. Add commands to the [flow] section of your Wavefile.")
		return 0
	}
	fmt.Fprintln(w, "Available flow commands:")
	for _, name := range cmds {
		fmt.Fprintf(w, "  %s\n", formatCommandListEntry(config, name))
	}
	return 0
}

func formatCommandListEntry(config map[string]any, name string) string {
	entryRaw, ok := config[name]
	if !ok {
		return name
	}
	entryMap, ok := entryRaw.(map[string]any)
	if !ok {
		return name
	}
	cmd, err := flow.ParseCommand(name, entryMap)
	if err != nil {
		return name
	}

	line := name
	if cmd.Desc != "" {
		line = fmt.Sprintf("%s - %s", line, cmd.Desc)
	}
	if len(cmd.Watch) > 0 {
		line = fmt.Sprintf("%s (watch: %s)", line, strings.Join(cmd.Watch, ", "))
	}
	return line
}
