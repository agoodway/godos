## Why

Todex now exposes first-class notes and goals APIs, while `godos` only uses the Todex API for todos and lists. Adding notes and goals commands makes `godos` a complete Todex CLI client and removes the remaining split-brain behavior where notes are still local filesystem data.

## What Changes

- **BREAKING** Replace local filesystem-backed note commands and note storage behavior with Todex API-backed notes.
- Add note folder resolution by user-facing folder name for note listing and creation.
- Add remote note commands for listing, creating, showing, editing, soft deleting, restoring, pinning, and unpinning notes.
- Add remote goal commands for listing, creating, showing, updating, deleting, and linking/unlinking tasks to goals.
- Display shortened UUID prefixes for notes and goals, and resolve user-provided prefixes to full UUIDs while rejecting missing or ambiguous matches.
- Refresh the vendored Todex OpenAPI contract from `/Users/tbrewer/projects/goodway/todex/openapi.json` and regenerate the Go client so note and goal operations are typed.
- Preserve existing Todex auth and configuration behavior; note and goal commands require the same configured base URL and bearer token as todo/list commands.

## Capabilities

### New Capabilities

- `remote-notes`: API-backed note folder and note behavior, including note listing/filtering, create/show/edit/delete/restore/pin/unpin commands, and note ID prefix resolution.
- `remote-goals`: API-backed goal behavior, including goal CRUD commands, progress display, task link/unlink commands, and goal ID prefix resolution.

### Modified Capabilities

- `notes`: Local filesystem note storage requirements are replaced by remote Todex note behavior.
- `cli-interface`: Existing note command requirements change from local note names and files to remote note IDs, note folders, and Todex-authenticated API behavior; new goal commands are added.

## Impact

- Affected commands: `notes`, `note add`, `note show`, `note edit`, `note rm`, plus new `note restore`, `note pin`, `note unpin`, `goals`, and `goal` subcommands.
- Affected packages: `cmd`, `internal/todex`, `internal/todexapi`, `api/openapi.json`, and tests.
- Existing local note storage in `internal/store` is no longer used by command behavior and may be removed or isolated as dead legacy code during implementation.
- Runtime behavior: notes and goals require network access and a valid Todex bearer token.
- Dependency behavior: generated OpenAPI client code must be regenerated from the updated Todex OpenAPI spec containing `/api/notes`, `/api/note-folders`, and `/api/goals` paths.
