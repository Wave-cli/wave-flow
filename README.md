# wave-flow

[![GitHub stars](https://img.shields.io/github/stars/Wave-cli/wave-flow?style=flat&logo=github)](https://github.com/Wave-cli/wave-flow/stargazers)
[![Issues](https://img.shields.io/github/issues/Wave-cli/wave-flow?style=flat&logo=github)](https://github.com/Wave-cli/wave-flow/issues)
[![License: MIT](https://img.shields.io/badge/license-MIT-brightgreen?style=flat)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.25.0-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/Wave-cli/wave-flow?style=flat&logo=github)](https://github.com/Wave-cli/wave-flow/releases)

Development workflow automation plugin for [wave](https://github.com/wave-cli/wave-core).

## Table of contents

- [What it does](#what-it-does)
- [Wavefile example](#wavefile-example)
- [Usage](#usage)
- [Install](#install)
- [Local development install](#local-development-install)
- [Config format](#config-format)
- [Errors](#errors)
- [Development](#development)

## What it does

wave-flow reads command definitions from the `[flow]` section of your Wavefile and executes them. Each command is an inline map with a required `cmd` field and optional callbacks and environment variables.

## Wavefile example

```toml
[project]
name = "my-app"
version = "1.0.0"

[flow]
build = { cmd = "go build -o bin/app", on_success = "echo done", env = { GOOS = "linux" } }
clean = { cmd = "rm -rf bin/" }
dev   = { cmd = "go run .", watch = ["*.go", "*.mod"] }
test  = { cmd = "go test ./...", on_fail = "echo tests failed" }
```

## Usage

```bash
# Run a command
wave flow build
wave flow clean

# List available commands
wave flow --list
```

## Install

```bash
wave install wave-cli/flow
```

## Local development install

Use this if you are working on `wave-flow` locally and want to test without publishing a release:

```bash
go build -o bin/flow .

mkdir -p ~/.wave/plugins/wave-cli/flow/v0.1.0/bin
cp bin/flow ~/.wave/plugins/wave-cli/flow/v0.1.0/bin/flow
cp Waveplugin ~/.wave/plugins/wave-cli/flow/v0.1.0/Waveplugin
ln -sfn ~/.wave/plugins/wave-cli/flow/v0.1.0 ~/.wave/plugins/wave-cli/flow/current
```

## Config format

Commands live under `[flow]` as inline tables. `cmd` is required.

- `cmd` (string, required): shell command to execute (`sh -c`)
- `on_success` (string, optional): runs if the main command exits 0
- `on_fail` (string, optional): runs if the main command exits non-zero
- `env` (table/map, optional): extra env vars (values coerced to strings)
- `watch` (string or array, optional): reserved for future watch mode

## Errors

wave-flow emits structured JSON errors on stderr using wave-core's SDK.

- Codes are lowercase-and-dashes (example: `flow-resolve-error`)
- Common codes: `flow-config-error`, `flow-no-command`, `flow-no-commands`, `flow-resolve-error`

## Development

```bash
just build    # Build binary
just test     # Run tests
just coverage # Generate coverage report
```

## License

MIT
