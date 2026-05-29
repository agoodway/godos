## 1. Storage Layer — Note Operations

- [x] 1.1 Add `NotesDir()` helper method to Store that returns the `notes/` subdirectory path
- [x] 1.2 Add `CreateNote(name string) error` method — validates name, creates `notes/<name>.md`, returns `ErrNoteExists` if duplicate
- [x] 1.3 Add `ReadNote(name string) (string, error)` method — returns file contents or `ErrNoteNotFound`
- [x] 1.4 Add `WriteNote(name string, content string) error` method — atomic write to `notes/<name>.md`
- [x] 1.5 Add `DeleteNote(name string) error` method — removes file or returns `ErrNoteNotFound`
- [x] 1.6 Add `ListNotes() ([]string, error)` method — scans `notes/*.md`, returns empty list if dir missing
- [x] 1.7 Add `NotePath(name string) (string, error)` helper with path traversal validation
- [x] 1.8 Add store tests for all note operations including error cases

## 2. CLI — Note Command Group

- [x] 2.1 Create `cmd/note.go` with `note` parent command and `add`, `show`, `edit`, `rm` subcommands
- [x] 2.2 Implement `note add <name>` — create note file and open in `$EDITOR` (default `vi`)
- [x] 2.3 Implement `note show <name>` — print note contents to stdout
- [x] 2.4 Implement `note edit <name>` — open existing note in `$EDITOR`
- [x] 2.5 Implement `note rm <name>` — delete with confirmation prompt, `--force` flag to skip
- [x] 2.6 Add integration tests for note subcommands

## 3. CLI — Notes List Command

- [x] 3.1 Create `cmd/notes.go` with `godos notes` command that lists all note names
- [x] 3.2 Add integration tests for notes listing and empty state
