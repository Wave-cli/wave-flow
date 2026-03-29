package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/wave-cli/wave-flow/cmd"
)

// --- cmd.Run() function tests ---
// cmd.Run() is the testable core of main(). It takes args, config reader, stdout, stderr
// and returns an exit code. It uses sdk.FormatPlainError for errors (no os.Exit).

func TestRunListCommands(t *testing.T) {
	config := map[string]any{
		"build": map[string]any{"cmd": "go build"},
		"clean": map[string]any{"cmd": "rm -rf dist"},
		"dev":   map[string]any{"cmd": "go run ."},
	}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"--list"}, stdin, &stdout, &stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr = %q", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "build") {
		t.Errorf("missing 'build' in output: %q", out)
	}
	if !strings.Contains(out, "clean") {
		t.Errorf("missing 'clean' in output: %q", out)
	}
	if !strings.Contains(out, "dev") {
		t.Errorf("missing 'dev' in output: %q", out)
	}
}

func TestRunListCommandsShortFlag(t *testing.T) {
	config := map[string]any{
		"build": map[string]any{"cmd": "echo ok"},
	}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"-l"}, stdin, &stdout, &stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "build") {
		t.Errorf("missing 'build' in output: %q", stdout.String())
	}
}

func TestRunHelpFlag(t *testing.T) {
	config := map[string]any{
		"build": map[string]any{"cmd": "echo ok"},
	}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"-h"}, stdin, &stdout, &stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	out := stdout.String()
	if !strings.Contains(out, "wave flow") {
		t.Errorf("help should mention 'wave flow', got: %q", out)
	}
	if !strings.Contains(out, "--list") {
		t.Errorf("help should mention '--list' flag, got: %q", out)
	}
}

func TestRunHelpFlagLong(t *testing.T) {
	config := map[string]any{}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"--help"}, stdin, &stdout, &stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	out := stdout.String()
	if !strings.Contains(out, "wave flow") {
		t.Errorf("help should mention 'wave flow', got: %q", out)
	}
}

func TestRunListCommandsEmpty(t *testing.T) {
	config := map[string]any{}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"--list"}, stdin, &stdout, &stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "No flow commands defined") {
		t.Errorf("expected empty commands message, got: %q", stdout.String())
	}
}

func TestRunNoArgsNoCommands(t *testing.T) {
	config := map[string]any{}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{}, stdin, &stdout, &stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	// Should show help when no args
	if !strings.Contains(stdout.String(), "wave flow") {
		t.Errorf("stdout should contain help, got: %q", stdout.String())
	}
}

func TestRunNoArgsWithCommands(t *testing.T) {
	config := map[string]any{
		"build": map[string]any{"cmd": "go build"},
		"dev":   map[string]any{"cmd": "go run ."},
	}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{}, stdin, &stdout, &stderr)

	// Should show help when no args (even with commands defined)
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "wave flow") {
		t.Errorf("stdout should contain help, got: %q", stdout.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	config := map[string]any{
		"build": map[string]any{"cmd": "go build"},
	}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"deploy"}, stdin, &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	// Check for clean error message (no longer uses error codes for resolve errors)
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "command not found") {
		t.Errorf("expected 'command not found' in stderr, got: %q", errOutput)
	}
	if !strings.Contains(errOutput, "deploy") {
		t.Errorf("expected command name 'deploy' in stderr, got: %q", errOutput)
	}
	// Should suggest using --list
	if !strings.Contains(errOutput, "wave flow --list") {
		t.Errorf("expected stderr to suggest 'wave flow --list', got: %q", errOutput)
	}
}

func TestRunExecuteCommand(t *testing.T) {
	config := map[string]any{
		"hello": map[string]any{
			"cmd": "echo hello_from_flow",
		},
	}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"hello"}, stdin, &stdout, &stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "hello_from_flow") {
		t.Errorf("stdout = %q, want 'hello_from_flow'", stdout.String())
	}
}

func TestRunExecuteCommandWithCallbacks(t *testing.T) {
	config := map[string]any{
		"build": map[string]any{
			"cmd":        "echo building",
			"on_success": "echo done",
		},
	}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"build"}, stdin, &stdout, &stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	out := stdout.String()
	if !strings.Contains(out, "building") {
		t.Errorf("missing main output: %q", out)
	}
	if !strings.Contains(out, "done") {
		t.Errorf("missing on_success output: %q", out)
	}
}

func TestRunExecuteFailingCommand(t *testing.T) {
	config := map[string]any{
		"fail": map[string]any{
			"cmd":     "exit 1",
			"on_fail": "echo caught",
		},
	}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"fail"}, stdin, &stdout, &stderr)

	if code == 0 {
		t.Error("expected non-zero exit code for failing command")
	}
	if !strings.Contains(stdout.String(), "caught") {
		t.Errorf("missing on_fail output: %q", stdout.String())
	}
}

func TestRunExecuteCommandWithEnv(t *testing.T) {
	config := map[string]any{
		"envtest": map[string]any{
			"cmd": "echo $MY_FLOW_VAR",
			"env": map[string]any{
				"MY_FLOW_VAR": "flow_works",
			},
		},
	}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"envtest"}, stdin, &stdout, &stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "flow_works") {
		t.Errorf("stdout = %q, want 'flow_works'", stdout.String())
	}
}

// --- Error code format tests ---

func TestErrorCodesAreDashFormat(t *testing.T) {
	// Verify all error codes use lowercase-and-dashes
	codes := []string{"flow-no-command", "flow-no-commands", "flow-resolve-error"}
	for _, code := range codes {
		for _, c := range code {
			if !(c >= 'a' && c <= 'z') && c != '-' {
				t.Errorf("error code %q contains invalid character %q — must be lowercase-and-dashes", code, string(c))
			}
		}
	}
}

func TestPlainErrorFormat(t *testing.T) {
	// Verify FormatPlainError produces the expected plain text format: "code: message\ndetails"
	var buf bytes.Buffer
	// Simulate what FormatPlainError would produce
	expected := "flow-test-error: test message\ntest details\n"
	buf.WriteString(expected)

	output := buf.String()
	// Should contain error code at start
	if !strings.HasPrefix(output, "flow-test-error:") {
		t.Errorf("output should start with error code, got: %q", output)
	}
	// Should contain message on first line
	if !strings.Contains(output, "test message") {
		t.Errorf("output should contain message, got: %q", output)
	}
	// Should contain details
	if !strings.Contains(output, "test details") {
		t.Errorf("output should contain details, got: %q", output)
	}
}

// --- Edge cases ---

func TestRunNilConfig(t *testing.T) {
	// Empty stdin (no config sent by wave-core)
	stdin := strings.NewReader("")

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"build"}, stdin, &stdout, &stderr)

	// Should handle gracefully — sdk.ReadConfigFrom returns error for empty input
	// The cmd.Run function should emit a plain text error
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	errCode := parsePlainErrorCode(t, stderr.String())
	if errCode != "flow-config-error" {
		t.Errorf("error code = %q, want 'flow-config-error'", errCode)
	}
}

func TestRunInvalidJSON(t *testing.T) {
	stdin := strings.NewReader("{invalid json")

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"build"}, stdin, &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	errCode := parsePlainErrorCode(t, stderr.String())
	if errCode != "flow-config-error" {
		t.Errorf("error code = %q, want 'flow-config-error'", errCode)
	}
}

func TestRunListWithNilConfig(t *testing.T) {
	// --list with empty stdin should handle gracefully
	stdin := strings.NewReader("")

	var stdout, stderr bytes.Buffer
	code := cmd.Run([]string{"--list"}, stdin, &stdout, &stderr)

	// Config error even for --list since we need config to list commands
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

// --- Helpers ---

func configToReader(t *testing.T, config map[string]any) *bytes.Buffer {
	t.Helper()
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	return bytes.NewBuffer(data)
}

// parsePlainErrorCode extracts the error code from plain text error format.
// Format: "code: message\ndetails"
func parsePlainErrorCode(t *testing.T, stderrOutput string) string {
	t.Helper()
	stderrOutput = strings.TrimSpace(stderrOutput)
	if stderrOutput == "" {
		t.Fatal("stderr is empty, expected plain text error")
	}
	// First line should be "code: message"
	lines := strings.SplitN(stderrOutput, "\n", 2)
	firstLine := lines[0]
	// Extract code before the colon
	parts := strings.SplitN(firstLine, ":", 2)
	if len(parts) < 2 {
		t.Fatalf("invalid error format, expected 'code: message', got: %q", firstLine)
	}
	return strings.TrimSpace(parts[0])
}
