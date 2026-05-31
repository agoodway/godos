## 1. OpenAPI Client Refresh

- [x] 1.1 Copy the updated Todex OpenAPI contract from `/Users/tbrewer/projects/goodway/todex/openapi.json` to `api/openapi.json`
- [x] 1.2 Regenerate `internal/todexapi/client.gen.go` using the existing `go generate` setup
- [x] 1.3 Verify generated client includes note folder, note, and goal schemas and operations
- [x] 1.4 Run existing todo/list/auth tests after regeneration and fix generated-client integration breaks

## 2. Remote Note Service

- [x] 2.1 Add command-facing note and note folder model types to `internal/todex`
- [x] 2.2 Add note folder list and folder-name resolution helpers, including default folder selection
- [x] 2.3 Add `ListNotes` service behavior with folder, query, pinned, and deleted filters
- [x] 2.4 Add `CreateNote` service behavior that resolves folder names to folder UUIDs
- [x] 2.5 Add note ID prefix resolution with missing and ambiguous prefix errors
- [x] 2.6 Add `GetNote`, `UpdateNoteBody`, `DeleteNote`, `RestoreNote`, `PinNote`, and `UnpinNote` service behavior
- [x] 2.7 Add service tests for note folder resolution, note filters, CRUD, lifecycle actions, and prefix errors

## 3. Remote Goal Service

- [x] 3.1 Add command-facing goal model types to `internal/todex`
- [x] 3.2 Add `ListGoals`, `CreateGoal`, `GetGoal`, `UpdateGoal`, and `DeleteGoal` service behavior
- [x] 3.3 Add goal ID prefix resolution with missing and ambiguous prefix errors
- [x] 3.4 Add `LinkGoalTask` and `UnlinkGoalTask` service behavior that resolves goal and task prefixes to full UUIDs
- [x] 3.5 Add service tests for goal CRUD, progress display data, task link/unlink, and prefix errors

## 4. Note Commands

- [x] 4.1 Add or replace `godos notes` to list remote notes with short IDs and filter flags
- [x] 4.2 Add or replace `godos note add <title> [--folder <name>]` to create a remote note and open the editor for body content
- [x] 4.3 Add or replace `godos note show <note-id-prefix>` to print remote note body content
- [x] 4.4 Add or replace `godos note edit <note-id-prefix>` using a temporary editor file and remote body update
- [x] 4.5 Add or replace `godos note rm <note-id-prefix> [--force]` with confirmation and remote soft delete
- [x] 4.6 Add `godos note restore`, `godos note pin`, and `godos note unpin` commands
- [x] 4.7 Ensure note commands use `getAPIService(true)` and no longer call `getStore()` for command behavior
- [x] 4.8 Add command tests for note list filters, add/edit editor flow, show, remove confirmation, restore, pin, unpin, and error cases

## 5. Goal Commands

- [x] 5.1 Add `godos goals` to list remote goals with short IDs, titles, and progress
- [x] 5.2 Add `godos goal add <title>` with `--description` and `--reason` flags
- [x] 5.3 Add `godos goal show <goal-id-prefix>` to display goal details
- [x] 5.4 Add `godos goal edit <goal-id-prefix>` with `--title`, `--description`, and `--reason` flags
- [x] 5.5 Add `godos goal rm <goal-id-prefix> [--force]` with confirmation and remote deletion
- [x] 5.6 Add `godos goal link <goal-id-prefix> <task-id-prefix>` and `godos goal unlink <goal-id-prefix> <task-id-prefix>` commands
- [x] 5.7 Add command tests for goal list, add, show, edit, remove, link, unlink, and error cases

## 6. Legacy Local Note Cleanup

- [x] 6.1 Remove local note command behavior from active command paths
- [x] 6.2 Remove or isolate obsolete local note storage tests that only validate removed command-facing behavior
- [x] 6.3 Keep any remaining `internal/store` code focused on still-used behavior, or remove note-only helpers if no code depends on them

## 7. Verification

- [x] 7.1 Run `openspec validate --all` and fix spec issues
- [x] 7.2 Run `go test ./...` and fix failures
- [x] 7.3 Run project quality checks if available
- [x] 7.4 Manually smoke test note and goal command flows against a local Todex API when available
