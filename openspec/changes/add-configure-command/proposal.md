## Why

godos currently has no way to persist user preferences. Users need a `configure` command to set defaults — starting with a default directory for file storage — so they don't have to specify paths on every invocation.

## What Changes

- Add a `godos configure` subcommand that reads and writes a persistent config file
- Support setting a `default_dir` configuration value that other commands can use as the default file storage location
- Store configuration in a platform-appropriate location (e.g., `~/.config/godos/config.yaml`)
- Support `godos configure set <key> <value>` and `godos configure get <key>` subcommands
- Support `godos configure list` to show all current settings

## Capabilities

### New Capabilities
- `configuration`: Persistent user configuration management — reading, writing, and resolving config values from a YAML config file

### Modified Capabilities
<!-- No existing specs to modify -->

## Impact

- New `cmd/configure.go` with Cobra subcommands
- New `config/` package for config file I/O
- New dependency: YAML library (e.g., `gopkg.in/yaml.v3`)
- Config file location: `~/.config/godos/config.yaml` (respects `XDG_CONFIG_HOME`)
