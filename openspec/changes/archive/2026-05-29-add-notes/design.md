## Context

godos stores todos as markdown checkbox lines in `.md` files under a storage directory (default `~/.godos/`). Users want a way to capture longer-form content — notes, ideas, references — alongside their todos. Notes are a distinct concept from todos: freeform markdown documents rather than single-line checklist items.

## Goals / Non-Goals

**Goals:**
- Provide note CRUD (create, read, update, delete) through CLI commands
- Store each note as a standalone `.md` file in a `notes/` subdirectory
- Support editing notes with the user's preferred editor (`$EDITOR`)
- Follow the same patterns as existing code (name validation, path safety, atomic writes)

**Non-Goals:**
- Linking notes to specific todos (notes and todos are independent)
- Note search or full-text search capabilities
- Tags, categories, or metadata on notes
- Syncing or sharing notes

## Decisions

### 1. Storage: `notes/` subdirectory with one file per note

Notes live at `<storage-dir>/notes/<name>.md`. The `notes/` subdirectory keeps them cleanly separated from todo list files in the storage root.

**Why this approach:**
- Simple filesystem layout — each note is a readable `.md` file
- No parser changes needed — notes are just files, not parsed structures
- Users can edit notes outside of godos if they want
- Name validation reuses existing `ValidateName` rules

**Alternatives considered:**
- Single `notes.md` file with sections — harder to manage individual notes, conflicts on concurrent edits
- Notes inline in todo files — couples two independent concepts, complicates the parser

### 2. Editing: delegate to `$EDITOR`

The `note edit` command opens the note file in the user's `$EDITOR` (falling back to `vi`). This avoids building an inline editor and gives users their preferred editing experience.

**Why:**
- Notes can be multi-line — capturing them as CLI arguments is awkward
- Users already have editor preferences
- Follows the pattern of `git commit` and other CLI tools

### 3. CLI structure: `note` command group + `notes` list

- `godos note add <name>` — create a new note (opens editor)
- `godos note show <name>` — print note contents to stdout
- `godos note edit <name>` — open existing note in editor
- `godos note rm <name>` — delete a note (with confirmation)
- `godos notes` — list all notes with summary info

This mirrors the `lists` (noun, list view) / list operations pattern already in godos.

### 4. Note creation: open editor immediately

When a user runs `note add <name>`, the file is created and the editor is opened so they can immediately write content. This reduces friction compared to a two-step create-then-edit flow.

## Risks / Trade-offs

- **[Editor dependency]** Users without `$EDITOR` set get `vi`, which may not be familiar → Mitigation: document the `EDITOR` env var; `vi` is universally available on Unix systems
- **[No Windows support for editor]** `$EDITOR` pattern is Unix-centric → Acceptable: godos uses `syscall.Flock` which is already Unix-only
- **[Name collisions with future features]** The `notes/` directory name is reserved → Low risk: clear naming, unlikely to conflict
