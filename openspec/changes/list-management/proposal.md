## Why

The base `godos` CLI creates lists implicitly via the `--list` flag, but provides no way to see what lists exist, rename them, or delete them. Users accumulate lists over time and need first-class commands to manage the lists themselves — not just the todos within them.

## What Changes

- New `godos lists` command to show all available lists with todo counts
- New `godos lists create <name>` subcommand to create an empty list
- New `godos lists rename <old> <new>` subcommand to rename a list (renames the `.md` file)
- New `godos lists delete <name>` subcommand to delete a list with confirmation prompt

## Capabilities

### New Capabilities

- `list-management`: Commands for creating, listing, renaming, and deleting todo lists

### Modified Capabilities

- `cli-interface`: Adding the `lists` command group to the CLI command tree

## Impact

- New `cmd/lists.go` file with the `lists` command and subcommands
- Extends `internal/store/` with rename and delete operations on list files
- No breaking changes — all existing commands continue to work unchanged
