package flow

import (
	"bytes"
	"strings"
	"testing"
)

// --- ParseCommand ---

func TestParseCommandHappyPath(t *testing.T) {
	entry := map[string]any{
		"cmd":         "go build -o ./bin/app",
		"description": "Build the app",
		"on_success":  "echo build done",
		"on_fail":     "echo build failed",
		"env": map[string]any{
			"GOOS":   "linux",
			"GOARCH": "amd64",
		},
	}

	cmd, err := ParseCommand("build", entry)
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}
	if cmd.Name != "build" {
		t.Errorf("Name = %q, want %q", cmd.Name, "build")
	}
	if cmd.Cmd != "go build -o ./bin/app" {
		t.Errorf("Cmd = %q", cmd.Cmd)
	}
	if cmd.OnSuccess != "echo build done" {
		t.Errorf("OnSuccess = %q", cmd.OnSuccess)
	}
	if cmd.OnFail != "echo build failed" {
		t.Errorf("OnFail = %q", cmd.OnFail)
	}
	if cmd.Desc != "Build the app" {
		t.Errorf("Desc = %q", cmd.Desc)
	}
	if cmd.Env["GOOS"] != "linux" {
		t.Errorf("Env[GOOS] = %q", cmd.Env["GOOS"])
	}
	if cmd.Env["GOARCH"] != "amd64" {
		t.Errorf("Env[GOARCH] = %q", cmd.Env["GOARCH"])
	}
}

func TestParseCommandMinimal(t *testing.T) {
	entry := map[string]any{
		"cmd": "echo hello",
	}

	cmd, err := ParseCommand("hello", entry)
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}
	if cmd.Cmd != "echo hello" {
		t.Errorf("Cmd = %q", cmd.Cmd)
	}
	if cmd.OnSuccess != "" {
		t.Errorf("OnSuccess should be empty, got %q", cmd.OnSuccess)
	}
	if cmd.OnFail != "" {
		t.Errorf("OnFail should be empty, got %q", cmd.OnFail)
	}
	if cmd.Desc != "" {
		t.Errorf("Desc should be empty, got %q", cmd.Desc)
	}
	if len(cmd.Env) != 0 {
		t.Errorf("Env should be empty, got %v", cmd.Env)
	}
}

func TestParseCommandDescAlias(t *testing.T) {
	entry := map[string]any{
		"cmd":  "echo hello",
		"desc": "Say hello",
	}

	cmd, err := ParseCommand("hello", entry)
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}
	if cmd.Desc != "Say hello" {
		t.Errorf("Desc = %q", cmd.Desc)
	}
}

func TestParseCommandMissingCmd(t *testing.T) {
	entry := map[string]any{
		"on_success": "echo done",
	}

	_, err := ParseCommand("build", entry)
	if err == nil {
		t.Fatal("Expected error for missing 'cmd' field")
	}
}

func TestParseCommandCmdNotString(t *testing.T) {
	entry := map[string]any{
		"cmd": 123,
	}

	_, err := ParseCommand("build", entry)
	if err == nil {
		t.Fatal("Expected error for non-string 'cmd' field")
	}
}

func TestParseCommandNilEntry(t *testing.T) {
	_, err := ParseCommand("build", nil)
	if err == nil {
		t.Fatal("Expected error for nil entry")
	}
}

func TestParseCommandEmptyEntry(t *testing.T) {
	_, err := ParseCommand("build", map[string]any{})
	if err == nil {
		t.Fatal("Expected error for empty entry (no cmd)")
	}
}

func TestParseCommandWatchString(t *testing.T) {
	entry := map[string]any{
		"cmd":   "go build",
		"watch": "*.go",
	}

	cmd, err := ParseCommand("build", entry)
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}
	if len(cmd.Watch) != 1 || cmd.Watch[0] != "*.go" {
		t.Errorf("Watch = %v, want [*.go]", cmd.Watch)
	}
}

func TestParseCommandWatchArray(t *testing.T) {
	entry := map[string]any{
		"cmd":   "go build",
		"watch": []any{"*.go", "*.mod"},
	}

	cmd, err := ParseCommand("build", entry)
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}
	if len(cmd.Watch) != 2 {
		t.Fatalf("Watch len = %d, want 2", len(cmd.Watch))
	}
	if cmd.Watch[0] != "*.go" || cmd.Watch[1] != "*.mod" {
		t.Errorf("Watch = %v", cmd.Watch)
	}
}

func TestParseCommandEnvValuesCoerced(t *testing.T) {
	entry := map[string]any{
		"cmd": "echo test",
		"env": map[string]any{
			"PORT":  float64(8080), // JSON numbers come as float64
			"DEBUG": true,
		},
	}

	cmd, err := ParseCommand("test", entry)
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}
	// Env values should be coerced to strings
	if cmd.Env["PORT"] != "8080" {
		t.Errorf("Env[PORT] = %q, want %q", cmd.Env["PORT"], "8080")
	}
	if cmd.Env["DEBUG"] != "true" {
		t.Errorf("Env[DEBUG] = %q, want %q", cmd.Env["DEBUG"], "true")
	}
}

// --- ResolveCommand ---

func TestResolveCommandFromConfig(t *testing.T) {
	config := map[string]any{
		"build": map[string]any{
			"cmd": "go build",
		},
		"clean": map[string]any{
			"cmd": "rm -rf dist",
		},
	}

	cmd, err := ResolveCommand(config, "build")
	if err != nil {
		t.Fatalf("ResolveCommand failed: %v", err)
	}
	if cmd.Cmd != "go build" {
		t.Errorf("Cmd = %q", cmd.Cmd)
	}
}

func TestResolveCommandNotFound(t *testing.T) {
	config := map[string]any{
		"build": map[string]any{
			"cmd": "go build",
		},
	}

	_, err := ResolveCommand(config, "deploy")
	if err == nil {
		t.Fatal("Expected error for unknown command")
	}
}

func TestResolveCommandEmptyConfig(t *testing.T) {
	_, err := ResolveCommand(map[string]any{}, "build")
	if err == nil {
		t.Fatal("Expected error for empty config")
	}
}

func TestResolveCommandNilConfig(t *testing.T) {
	_, err := ResolveCommand(nil, "build")
	if err == nil {
		t.Fatal("Expected error for nil config")
	}
}

func TestResolveCommandEntryNotMap(t *testing.T) {
	config := map[string]any{
		"build": "not a map",
	}

	_, err := ResolveCommand(config, "build")
	if err == nil {
		t.Fatal("Expected error when command entry is not a map")
	}
}

// --- ListCommands ---

func TestListCommands(t *testing.T) {
	config := map[string]any{
		"build": map[string]any{"cmd": "go build"},
		"clean": map[string]any{"cmd": "rm -rf dist"},
		"dev":   map[string]any{"cmd": "go run ."},
	}

	cmds := ListCommands(config)
	if len(cmds) != 3 {
		t.Fatalf("ListCommands returned %d, want 3", len(cmds))
	}

	// Should be sorted
	if cmds[0] != "build" || cmds[1] != "clean" || cmds[2] != "dev" {
		t.Errorf("ListCommands = %v, want [build clean dev]", cmds)
	}
}

func TestListCommandsEmpty(t *testing.T) {
	cmds := ListCommands(map[string]any{})
	if len(cmds) != 0 {
		t.Errorf("ListCommands for empty config should be empty, got %v", cmds)
	}
}

func TestListCommandsNil(t *testing.T) {
	cmds := ListCommands(nil)
	if len(cmds) != 0 {
		t.Errorf("ListCommands for nil config should be empty, got %v", cmds)
	}
}

// --- RunCommand ---

func TestRunCommandSimple(t *testing.T) {
	cmd := &Command{
		Name: "hello",
		Cmd:  "echo hello world",
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCommand(cmd, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("ExitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "hello world") {
		t.Errorf("Stdout = %q, expected to contain 'hello world'", stdout.String())
	}
}

func TestRunCommandWithEnv(t *testing.T) {
	cmd := &Command{
		Name: "envtest",
		Cmd:  "echo $WAVE_TEST_VAR",
		Env:  map[string]string{"WAVE_TEST_VAR": "it_works"},
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCommand(cmd, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("ExitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "it_works") {
		t.Errorf("Stdout = %q, expected to contain 'it_works'", stdout.String())
	}
}

func TestRunCommandFailure(t *testing.T) {
	cmd := &Command{
		Name: "fail",
		Cmd:  "exit 1",
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCommand(cmd, &stdout, &stderr)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for failing command")
	}
}

func TestRunCommandWithOnSuccess(t *testing.T) {
	cmd := &Command{
		Name:      "build",
		Cmd:       "echo main",
		OnSuccess: "echo callback",
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCommand(cmd, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("ExitCode = %d", exitCode)
	}
	output := stdout.String()
	if !strings.Contains(output, "main") {
		t.Errorf("Missing main output: %q", output)
	}
	if !strings.Contains(output, "callback") {
		t.Errorf("Missing on_success callback output: %q", output)
	}
}

func TestRunCommandWithOnFail(t *testing.T) {
	cmd := &Command{
		Name:   "build",
		Cmd:    "exit 1",
		OnFail: "echo failure_callback",
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCommand(cmd, &stdout, &stderr)

	// Exit code should still be non-zero (from the main command)
	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
	output := stdout.String()
	if !strings.Contains(output, "failure_callback") {
		t.Errorf("Missing on_fail callback output: %q", output)
	}
}

func TestRunCommandOnSuccessNotCalledOnFailure(t *testing.T) {
	cmd := &Command{
		Name:      "build",
		Cmd:       "exit 1",
		OnSuccess: "echo should_not_appear",
	}

	var stdout, stderr bytes.Buffer
	RunCommand(cmd, &stdout, &stderr)

	if strings.Contains(stdout.String(), "should_not_appear") {
		t.Error("on_success should not be called when main command fails")
	}
}

func TestRunCommandOnFailNotCalledOnSuccess(t *testing.T) {
	cmd := &Command{
		Name:   "build",
		Cmd:    "echo ok",
		OnFail: "echo should_not_appear",
	}

	var stdout, stderr bytes.Buffer
	RunCommand(cmd, &stdout, &stderr)

	if strings.Contains(stdout.String(), "should_not_appear") {
		t.Error("on_fail should not be called when main command succeeds")
	}
}

func TestRunCommandEmptyCmd(t *testing.T) {
	cmd := &Command{
		Name: "empty",
		Cmd:  "",
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCommand(cmd, &stdout, &stderr)

	// Empty command should result in an error
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for empty command")
	}
}

func TestRunCommandWithSpecialCharacters(t *testing.T) {
	cmd := &Command{
		Name: "special",
		Cmd:  `echo "hello 'world' & $PATH"`,
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCommand(cmd, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("ExitCode = %d, stderr = %q", exitCode, stderr.String())
	}
}

func TestRunCommandWithPipe(t *testing.T) {
	cmd := &Command{
		Name: "pipe",
		Cmd:  "echo 'hello world' | tr 'h' 'H'",
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCommand(cmd, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("ExitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Hello world") {
		t.Errorf("Stdout = %q, expected to contain 'Hello world'", stdout.String())
	}
}
