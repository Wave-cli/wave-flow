# wave-flow Quick Start

wave-flow is a workflow automation plugin for wave. It lets you define and run project commands from your Wavefile.

## Installation

```bash
wave install wave-cli/flow
```

## Basic Usage

### 1. Define commands in your Wavefile

```toml
[flow]
build = { cmd = "go build -o bin/app" }
dev   = { cmd = "go run ." }
test  = { cmd = "go test ./..." }
```

### 2. Run commands

```bash
wave flow build
wave flow dev
wave flow test
```

### 3. List available commands

```bash
wave flow --list
```

## Command Options

| Option | Description |
|--------|-------------|
| `cmd` | Shell command to run (required) |
| `env` | Environment variables as key-value pairs |
| `on_success` | Command to run if main command succeeds |
| `on_fail` | Command to run if main command fails |
| `watch` | File patterns for watch mode (future) |

## Example with all options

```toml
[flow.build]
cmd = "go build -o bin/app"
env = { CGO_ENABLED = "0", GOOS = "linux" }
on_success = "echo Build complete"
on_fail = "echo Build failed"
```

## Help

```bash
wave flow --help
```
