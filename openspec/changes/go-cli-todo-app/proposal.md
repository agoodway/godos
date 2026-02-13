## Why

A simple, fast CLI tool for managing todos directly in markdown files. Markdown is portable, version-controllable, and human-readable without specialized tools. Existing todo apps either require a GUI, use proprietary formats, or add unnecessary complexity. A Go CLI keeps things fast, single-binary, and scriptable.

## What Changes

- New Go CLI application (`godos`) for managing todos
- Todos stored as markdown files — one file per list, tasks as `- [ ]` / `- [x]` checkboxes
- Commands for adding, completing, listing, and removing todos
- Support for multiple named lists (default list if none specified)
- Configurable storage directory (defaults to `~/.godos/`)

## Capabilities

### New Capabilities

- `cli-interface`: Command parsing, flags, help text, and user-facing output formatting
- `todo-storage`: Reading, writing, and parsing markdown files as todo lists
- `todo-operations`: Core CRUD operations — add, complete, remove, list todos

### Modified Capabilities

_None — greenfield project._

## Impact

- New Go module and binary (`godos`)
- Creates markdown files in a configurable directory on the user's filesystem
- No external dependencies beyond Go stdlib (or minimal CLI library)
