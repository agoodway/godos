## Context

The base `go-cli-todo-app` change provides implicit list creation (writing to a named `.md` file). This change adds explicit list management commands. The storage layer already handles list discovery (`*.md` scan) and atomic file writes.

## Goals / Non-Goals

**Goals:**
- First-class `lists` command group for managing lists
- Show list inventory with todo counts at a glance
- Safe delete with confirmation to prevent accidental data loss

**Non-Goals:**
- List metadata (descriptions, created dates)
- List archiving or soft-delete
- List templates or copying

## Decisions

### 1. Command structure: `godos lists [subcommand]`
**Rationale**: `godos lists` (no subcommand) shows all lists — the most common operation. Subcommands `create`, `rename`, `delete` handle mutations. This follows the `kubectl get/create/delete` pattern. Alternative considered: top-level commands (`godos create-list`) — clutters the main command namespace.

### 2. Delete confirmation: `--force` flag to skip
**Rationale**: `godos lists delete <name>` prompts "Delete list '<name>' with N todos? [y/N]" by default. `--force` skips the prompt for scripting. Alternative considered: no confirmation — too risky since delete is destructive.

### 3. Rename implementation: file rename
**Rationale**: Renaming a list is simply renaming `<old>.md` to `<new>.md` in the storage directory. No content changes needed. The storage layer handles this with `os.Rename`.

### 4. New files
```
cmd/lists.go        # lists command group + create/rename/delete subcommands
```
Plus extensions to `internal/store/store.go` for `Rename()` and `Delete()` methods.

## Risks / Trade-offs

- [Rename while other process reads] → Same low risk as all file ops in this single-user CLI. Atomic rename is as safe as it gets.
- [Delete is permanent] → Mitigated by confirmation prompt. Users can also use git to recover if their todo dir is versioned.
