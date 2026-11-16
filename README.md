# Falcula

A programmable toolkit for services, containers, and image generation.

Falcula is a tool that can be used to create and manage services, containers, and image generation. It works by running a Lua script in the
background and providing a TUI (optional) to control it. Falcula exposes a set of Lua modules to create and manage services and containers.

## Installation

Falcula CLI can be installed with the following command:

```sh
go install github.com/LucasAVasco/falcula/cmd/falcula@latest
```

Falcula uses [Gopher Lua](https://github.com/yuin/gopher-lua) to run Lua scripts, so you do not need to install a Lua runtime to use
Falcula.

### Optional dependencies

Some Falcula modules require optional dependencies. You need to install them before using the corresponding module.

- 'falcula.docker.compose' depends on [Docker](https://www.docker.com/). This module use the native docker CLI to run docker-compose
  (`docker compose` command) instead of the `docker-compose` binary. The 'falcula.compose' is a alias for the 'falcula.docker.compose'
  module.

## Usage

To start Falcula, run the following command:

```sh
falcula run [arguments...]
```

This command will open the TUI and run the Lua script in the background with the given arguments.

If you want to run the script without the TUI, use the `--raw` flag:

```sh
falcula --raw run [arguments...]
```

Falcula also runs without the TUI automatically when it can not use a terminal: in CI environments (`CI` or a CI provider variable is
set), when the standard output is redirected (e.g. `falcula run task > log.txt` or `falcula run task | less`) or when the standard input
is not a terminal.

### Shell completion

`falcula run` (and `falcula task run`) completes task names, including the tasks of sub-projects (`project1:subproject2:my-task`).
Only the projects in the typed name are read, so completion stays fast in large project trees. Enable it in your shell:

```sh
source <(falcula completion bash)          # Bash (add to ~/.bashrc)
source <(falcula completion zsh)           # Zsh (add to ~/.zshrc)
falcula completion fish | source           # Fish (add to ~/.config/fish/config.fish)
```

See `falcula completion --help` for other shells and for installing the scripts permanently.

### Opening service files

In the TUI, press `o` on a service to open its associated file in your editor (`$VISUAL`, `$EDITOR` or `vi`). Docker compose services
open their compose file. Any service can set its own file with the `file` option (relative to the Lua file that creates it):

```lua
local api = process_provider:new_service({ file = "api/main.go" })
```

On a manager, `o` opens the files of all its services (each file once).

## Configuration and themes

Falcula reads its user configuration from `$FALCULA_CONFIG_DIR`, or from the `falcula` folder in the user configuration folder
(`~/.config/falcula` on Linux, honoring `$XDG_CONFIG_HOME`). Nothing is required: without configuration, falcula uses the built-in `soft`
theme (soft, non-vibrant colors on the terminal's own background).

```
~/.config/falcula/
  config.yaml
  themes/
    soft.yaml        # Optional copy of the built-in theme
    my-theme.yaml    # Your themes (selected by the file name without the extension)
```

Run `falcula config init` to write `config.yaml` and `themes/soft.yaml` as a starting point (existing files are kept unless `--force` is
used).

`config.yaml`:

```yaml
theme: my-theme # File name in the themes folder (without the extension) or a built-in theme (soft)
```

A theme only needs the keys it changes. The missing ones come from the theme in `extends` (default: `soft`). Colors are `"#rrggbb"`, a color
name (e.g. `slategray`) or `default` (the terminal's own color):

```yaml
extends: soft
colors:
  border_focused: "#b48ead"
  selection_background: "#b48ead"
status:
  running: "#88c0a0"
```

See `themes/soft.yaml` (written by `falcula config init`) for all keys. If the configuration or the theme has an error, falcula uses the
default theme and shows the error in the "Debug logs" panel of the TUI.

## Running tests

The unit tests only need Go. The integration tests run real processes and the `falcula` CLI. They use a fake `docker` executable, so Docker is not
required. They are behind the `integration` build tag.

```sh
go test ./...                                                # Unit tests (or `make test`)
go test -tags integration -run '^TestIntegration' ./...      # Integration tests (or `make test-integration`)
go test -race -tags integration ./...                        # Everything with the race detector (or `make test-all`)
make cover                                                   # Coverage report
```
