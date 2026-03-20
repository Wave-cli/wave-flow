# wave-flow Command Reference

## Commands

### Run a flow command

```bash
wave flow <command>
```

Executes the specified command from your Wavefile's `[flow]` section.

### List commands

```bash
wave flow --list
wave flow -l
```

Shows all available flow commands defined in your Wavefile.

### Help

```bash
wave flow --help
wave flow -h
```

Displays usage information.

## Wavefile Configuration

### Basic command

```toml
[flow]
build = { cmd = "go build ./..." }
```

### Command with environment variables

```toml
[flow]
build = { cmd = "go build -o bin/app", env = { CGO_ENABLED = "0" } }
```

### Command with callbacks

```toml
[flow]
test = { cmd = "go test ./...", on_success = "echo passed", on_fail = "echo failed" }
```

### Full example

```toml
[flow]
build = { cmd = "go build -o bin/app" }
dev   = { cmd = "go run .", env = { PORT = "3000" } }
test  = { cmd = "go test ./...", on_fail = "echo Tests failed!" }
lint  = { cmd = "go vet ./..." }
clean = { cmd = "rm -rf bin/" }
```

## Error Codes

| Code | Description |
|------|-------------|
| `flow-config-error` | Failed to read Wavefile configuration |
| `flow-resolve-error` | Command not found in Wavefile |

Use `--debug` flag with wave to see detailed error information:

```bash
wave --debug flow build
```
