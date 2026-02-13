## Context

Greenfield Go CLI application. No existing codebase — building from scratch. The tool manages todos stored as markdown files with standard checkbox syntax (`- [ ]` / `- [x]`).

## Goals / Non-Goals

**Goals:**
- Single static binary with zero runtime dependencies
- Fast startup — CLI should feel instant
- Human-readable markdown storage that works with git
- Simple, memorable command interface

**Non-Goals:**
- GUI or TUI interface (pure CLI)
- Syncing or collaboration features
- Due dates, priorities, or tags (keep v1 minimal)
- Nested/hierarchical todos

## Decisions

### 1. CLI framework: `cobra`
**Rationale**: Cobra is the de facto standard for Go CLIs (used by kubectl, gh, hugo). Provides subcommand routing, flag parsing, and auto-generated help. Alternative considered: bare `flag` package — simpler but lacks subcommands and help generation.

### 2. Storage format: one markdown file per list
**Rationale**: Each list is a `.md` file in the storage directory. The filename (without extension) is the list name. Todos are `- [ ] text` or `- [x] text` lines. A default list (`todo.md`) is used when no list is specified.

Example `todo.md`:
```markdown
- [x] Set up Go module
- [ ] Write proposal
- [ ] Implement add command
```

### 3. Storage location: `~/.godos/` default, configurable via `--dir` flag
**Rationale**: Dotfile in home directory is conventional. Global `--dir` flag allows pointing at any directory (e.g., a project-local `.todos/` folder). No config file needed for v1.

### 4. Project structure
```
godos/
├── main.go           # Entry point
├── cmd/              # Cobra commands
│   ├── root.go       # Root command, global flags
│   ├── add.go        # godos add
│   ├── list.go       # godos list
│   ├── done.go       # godos done
│   └── remove.go     # godos rm
├── internal/
│   └── store/        # Markdown file I/O and parsing
│       ├── store.go  # Store type and operations
│       └── parse.go  # Markdown parsing/rendering
├── go.mod
└── go.sum
```

### 5. Todo identification: line number
**Rationale**: Todos are identified by their 1-based line number in the list. `godos done 3` marks the 3rd todo as complete. Simple and requires no ID generation. Alternative considered: text matching — fragile with duplicates.

## Risks / Trade-offs

- [Line-number IDs shift after adds/removes] → Acceptable for a CLI tool; user re-runs `godos list` to see current numbers. Matches `kill` and similar CLI conventions.
- [No concurrency protection on file writes] → Low risk for single-user CLI. File operations are atomic enough via write-temp-then-rename.
- [Cobra adds ~5MB to binary size] → Acceptable trade-off for ergonomic CLI experience.
