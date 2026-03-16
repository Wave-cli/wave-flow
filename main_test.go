package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/wave-cli/wave-core/pkg/sdk"
)

// --- run() function tests ---
// run() is the testable core of main(). It takes args, config reader, stdout, stderr
// and returns an exit code. It uses sdk.FormatWaveError for errors (no os.Exit).

func TestRunListCommands(t *testing.T) {
	config := map[string]any{
		"build": map[string]any{"cmd": "go build"},
		"clean": map[string]any{"cmd": "rm -rf dist"},
		"dev":   map[string]any{"cmd": "go run ."},
	}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := run([]string{"--list"}, stdin, &stdout, &stderr)

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
	code := run([]string{"-l"}, stdin, &stdout, &stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "build") {
		t.Errorf("missing 'build' in output: %q", stdout.String())
	}
}

func TestRunListCommandsEmpty(t *testing.T) {
	config := map[string]any{}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := run([]string{"--list"}, stdin, &stdout, &stderr)

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
	code := run([]string{}, stdin, &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	waveErr := parseWaveError(t, stderr.String())
	if waveErr.Code != "flow-no-commands" {
		t.Errorf("error code = %q, want 'flow-no-commands'", waveErr.Code)
	}
}

func TestRunNoArgsWithCommands(t *testing.T) {
	config := map[string]any{
		"build": map[string]any{"cmd": "go build"},
		"dev":   map[string]any{"cmd": "go run ."},
	}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := run([]string{}, stdin, &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	waveErr := parseWaveError(t, stderr.String())
	if waveErr.Code != "flow-no-command" {
		t.Errorf("error code = %q, want 'flow-no-command'", waveErr.Code)
	}
	if !strings.Contains(waveErr.Details, "build") {
		t.Errorf("details should list available commands, got: %q", waveErr.Details)
	}
	if !strings.Contains(waveErr.Details, "dev") {
		t.Errorf("details should list available commands, got: %q", waveErr.Details)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	config := map[string]any{
		"build": map[string]any{"cmd": "go build"},
	}
	stdin := configToReader(t, config)

	var stdout, stderr bytes.Buffer
	code := run([]string{"deploy"}, stdin, &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	waveErr := parseWaveError(t, stderr.String())
	if waveErr.Code != "flow-resolve-error" {
		t.Errorf("error code = %q, want 'flow-resolve-error'", waveErr.Code)
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
	code := run([]string{"hello"}, stdin, &stdout, &stderr)

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
	code := run([]string{"build"}, stdin, &stdout, &stderr)

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
	code := run([]string{"fail"}, stdin, &stdout, &stderr)

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
	code := run([]string{"envtest"}, stdin, &stdout, &stderr)

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

func TestWaveErrorStructure(t *testing.T) {
	// Verify FormatWaveError produces valid wave error JSON
	var buf bytes.Buffer
	sdk.FormatWaveError(&buf, "flow-test-error", "test message", "test details")

	var waveErr sdk.WaveError
	if err := json.Unmarshal(buf.Bytes(), &waveErr); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if !waveErr.WaveError {
		t.Error("wave_error should be true")
	}
	if waveErr.Code != "flow-test-error" {
		t.Errorf("code = %q", waveErr.Code)
	}
	if waveErr.Message != "test message" {
		t.Errorf("message = %q", waveErr.Message)
	}
	if waveErr.Details != "test details" {
		t.Errorf("details = %q", waveErr.Details)
	}
}

// --- Edge cases ---

func TestRunNilConfig(t *testing.T) {
	// Empty stdin (no config sent by wave-core)
	stdin := strings.NewReader("")

	var stdout, stderr bytes.Buffer
	code := run([]string{"build"}, stdin, &stdout, &stderr)

	// Should handle gracefully — sdk.ReadConfigFrom returns error for empty input
	// The run function should emit a wave error
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	waveErr := parseWaveError(t, stderr.String())
	if waveErr.Code != "flow-config-error" {
		t.Errorf("error code = %q, want 'flow-config-error'", waveErr.Code)
	}
}

func TestRunInvalidJSON(t *testing.T) {
	stdin := strings.NewReader("{invalid json")

	var stdout, stderr bytes.Buffer
	code := run([]string{"build"}, stdin, &stdout, &stderr)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	waveErr := parseWaveError(t, stderr.String())
	if waveErr.Code != "flow-config-error" {
		t.Errorf("error code = %q, want 'flow-config-error'", waveErr.Code)
	}
}

func TestRunListWithNilConfig(t *testing.T) {
	// --list with empty stdin should handle gracefully
	stdin := strings.NewReader("")

	var stdout, stderr bytes.Buffer
	code := run([]string{"--list"}, stdin, &stdout, &stderr)

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

func parseWaveError(t *testing.T, stderrOutput string) sdk.WaveError {
	t.Helper()
	var waveErr sdk.WaveError
	if err := json.Unmarshal([]byte(strings.TrimSpace(stderrOutput)), &waveErr); err != nil {
		t.Fatalf("failed to parse wave error from stderr: %v\nstderr = %q", err, stderrOutput)
	}
	if !waveErr.WaveError {
		t.Fatal("wave_error should be true")
	}
	return waveErr
}
