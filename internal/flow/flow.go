// Package flow implements the core logic for the wave-flow plugin.
// It parses command definitions from the Wavefile [flow] section,
// executes commands with optional environment variables, and handles
// on_success/on_fail callbacks.
package flow

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// Command represents a single flow command definition.
type Command struct {
	Name      string            // Command name (e.g. "build", "clean")
	Cmd       string            // Shell command to execute
	OnSuccess string            // Command to run on success (optional)
	OnFail    string            // Command to run on failure (optional)
	Env       map[string]string // Extra environment variables (optional)
	Watch     []string          // File patterns to watch (optional)
}

// ParseCommand parses a command definition from a raw config entry map.
func ParseCommand(name string, entry map[string]any) (*Command, error) {
	if entry == nil {
		return nil, fmt.Errorf("command %q: entry is nil", name)
	}

	// Extract required "cmd" field
	cmdRaw, ok := entry["cmd"]
	if !ok {
		return nil, fmt.Errorf("command %q: missing required 'cmd' field", name)
	}
	cmdStr, ok := cmdRaw.(string)
	if !ok {
		return nil, fmt.Errorf("command %q: 'cmd' must be a string, got %T", name, cmdRaw)
	}

	cmd := &Command{
		Name: name,
		Cmd:  cmdStr,
		Env:  make(map[string]string),
	}

	// Optional string fields
	if v, ok := entry["on_success"].(string); ok {
		cmd.OnSuccess = v
	}
	if v, ok := entry["on_fail"].(string); ok {
		cmd.OnFail = v
	}

	// Optional env map
	if envRaw, ok := entry["env"]; ok {
		if envMap, ok := envRaw.(map[string]any); ok {
			for k, v := range envMap {
				cmd.Env[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// Optional watch (string or array)
	if watchRaw, ok := entry["watch"]; ok {
		switch w := watchRaw.(type) {
		case string:
			cmd.Watch = []string{w}
		case []any:
			for _, item := range w {
				if s, ok := item.(string); ok {
					cmd.Watch = append(cmd.Watch, s)
				}
			}
		}
	}

	return cmd, nil
}

// ResolveCommand looks up a command by name in the config section and parses it.
func ResolveCommand(config map[string]any, name string) (*Command, error) {
	if config == nil {
		return nil, fmt.Errorf("no commands configured (empty config)")
	}

	entryRaw, ok := config[name]
	if !ok {
		available := ListCommands(config)
		return nil, fmt.Errorf("unknown command %q. Available: %s", name, strings.Join(available, ", "))
	}

	entryMap, ok := entryRaw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("command %q: expected a map, got %T", name, entryRaw)
	}

	return ParseCommand(name, entryMap)
}

// ListCommands returns sorted command names from a config section.
func ListCommands(config map[string]any) []string {
	if config == nil {
		return nil
	}
	names := make([]string, 0, len(config))
	for k := range config {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// RunCommand executes a flow command and its callbacks.
// Returns the exit code of the main command.
func RunCommand(cmd *Command, stdout, stderr io.Writer) int {
	if cmd.Cmd == "" {
		fmt.Fprintln(stderr, "error: empty command")
		return 1
	}

	// Execute main command
	exitCode := shellExec(cmd.Cmd, cmd.Env, stdout, stderr)

	// Run callbacks
	if exitCode == 0 && cmd.OnSuccess != "" {
		shellExec(cmd.OnSuccess, cmd.Env, stdout, stderr)
	}
	if exitCode != 0 && cmd.OnFail != "" {
		shellExec(cmd.OnFail, cmd.Env, stdout, stderr)
	}

	return exitCode
}

// shellExec runs a command string via sh -c, inheriting env + extra vars.
func shellExec(cmdStr string, extraEnv map[string]string, stdout, stderr io.Writer) int {
	c := exec.Command("sh", "-c", cmdStr)
	c.Stdout = stdout
	c.Stderr = stderr

	// Inherit environment + add extra vars
	env := os.Environ()
	for k, v := range extraEnv {
		env = append(env, k+"="+v)
	}
	c.Env = env

	if err := c.Run(); err != nil {
		if c.ProcessState != nil {
			return c.ProcessState.ExitCode()
		}
		return 1
	}

	return 0
}
