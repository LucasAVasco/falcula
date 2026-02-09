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

- 'falcula.docker-compose' depends on [Docker](https://www.docker.com/). This module use the native docker CLI to run docker-compose
  (`docker compose` command) instead of the `docker-compose` binary. The 'falcula.compose' is a alias for the 'falcula.docker-compose'
  module.

## Usage

To start Falcula, run the following command:

```sh
falcula run [arguments...]
```

This command will open the TUI and run the Lua script in the background with the given arguments.

If you want to run the script without the TUI (e.g. in a CI environment), use the following command:

```sh
falcula run-raw [arguments...]
```
