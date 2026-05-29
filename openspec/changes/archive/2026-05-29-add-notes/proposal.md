## Why

godos currently only supports todos — short checklist items. Users also need to capture longer-form thoughts, meeting notes, ideas, or reference material alongside their todos. Notes are a separate content type: freeform markdown files stored in a `notes/` subdirectory, managed through their own CLI commands.

## What Changes

- Add a `notes/` subdirectory inside the storage directory for note files
- Add a `godos note add <name>` command that creates a new note as a `.md` file
- Add a `godos note edit <name>` command that opens a note in the user's `$EDITOR`
- Add a `godos note show <name>` command that displays a note's content
- Add a `godos note rm <name>` command that removes a note file
- Add a `godos notes` command that lists all notes

## Capabilities

### New Capabilities
- `notes`: Creating, listing, viewing, editing, and removing notes stored as individual markdown files in a `notes/` subdirectory

### Modified Capabilities
- `cli-interface`: New `note` command group and `notes` list command

## Impact

- `internal/store/store.go`: New note methods operating on `notes/` subdirectory — file CRUD, no parser changes needed
- `cmd/`: New `note.go` and `notes.go` command files
- No changes to existing todo storage or parsing — notes are entirely separate files
