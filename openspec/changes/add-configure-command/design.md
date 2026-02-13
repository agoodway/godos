## Context

godos is a Go CLI built with Cobra. It currently has `root` and `version` commands. There is no configuration system — all behavior is determined by flags at invocation time. The user wants a `configure` command so preferences like a default storage directory persist across sessions.

## Goals / Non-Goals

**Goals:**
- Provide a `godos configure` command with `set`, `get`, and `list` subcommands
- Persist config in a YAML file at a platform-appropriate location
- Expose a `config` package that other commands can import to resolve settings

**Non-Goals:**
- Environment variable overrides (future work)
- Per-project config files (only user-level config for now)
- Config file encryption or secrets management

## Decisions

### Config file format: YAML
**Rationale**: YAML is human-readable, well-supported in Go (`gopkg.in/yaml.v3`), and standard for CLI config files. JSON is noisier for hand-editing; TOML has weaker Go ecosystem support.

### Config file location: `~/.config/godos/config.yaml`
**Rationale**: Follows XDG Base Directory spec. Respects `XDG_CONFIG_HOME` if set, falls back to `~/.config`. This is the convention for Unix CLI tools.

### Flat key-value model
**Rationale**: Start simple with a flat `map[string]string` internally. Keys are dotted strings (e.g., `default_dir`). Avoids premature nesting complexity. The YAML file will use top-level keys.

### Package structure: `config/config.go`
**Rationale**: Separate package so any command can `import "github.com/goodway/godos/config"` and call `config.Get("default_dir")` without coupling to the CLI layer.

### Command structure: `configure` with `set`, `get`, `list` subcommands
**Rationale**: Follows the pattern of `git config` and `aws configure`. Subcommands are clearer than flags for CRUD operations.

## Risks / Trade-offs

- [File permissions] Config file created with default permissions → Acceptable for non-sensitive settings. No secrets stored.
- [Directory creation] Config dir may not exist on first run → Auto-create `~/.config/godos/` with `os.MkdirAll` on first `set`.
- [Concurrent writes] No file locking → Acceptable for a single-user CLI tool. Race conditions are negligible.
