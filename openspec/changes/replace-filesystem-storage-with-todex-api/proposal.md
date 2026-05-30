## Why

`godos` currently stores todo lists as local markdown files, which prevents server-backed sync and requires filesystem-specific behavior throughout the CLI. The Todex API already exposes authenticated list and task operations, so `godos` should become an API-only client for that backend.

## What Changes

- **BREAKING** Replace filesystem-backed todo/list storage with the Todex HTTP API described by `/Users/tbrewer/projects/goodway/todex/openapi.json`.
- **BREAKING** Remove positional todo mutation arguments such as `godos done 3` and `godos rm 3`; task mutations SHALL require task ID prefixes.
- Add API authentication commands: `godos register` and `godos login`, both storing the returned JWT bearer token for future API requests.
- Add API configuration for the Todex base URL and bearer token.
- Preserve the existing user-facing todo/list workflow where practical: `add`, `list`, `done`, `rm`, and `lists` commands continue to exist.
- Auto-create a remote list when `godos add --list <name>` targets a missing list.
- Display shortened UUID prefixes for tasks and resolve user-provided prefixes to full task UUIDs, rejecting ambiguous prefixes.
- Keep notes out of scope for this change; existing note behavior is not migrated to the Todex API.

## Capabilities

### New Capabilities

- `todex-api-client`: Generated or hand-written client integration for the Todex OpenAPI contract, including authenticated list/task requests and error handling.
- `api-authentication`: User registration, login, JWT bearer token persistence, and authenticated API request behavior.
- `remote-todo-storage`: Remote list and task behavior that replaces local markdown-file todo storage.
- `task-id-prefixes`: Shortened UUID display and unique-prefix resolution for task mutation commands.

### Modified Capabilities

- `configuration`: `godos configure list` SHALL mask the stored Todex JWT bearer token instead of printing it verbatim.

### Removed Capabilities / Requirements

This change supersedes the filesystem-era todo behavior established by the (not-yet-archived) `go-cli-todo-app` and `list-management` changes. It therefore removes their now-contradictory requirements:

- `cli-interface`: removes the filesystem/positional `Add`, `List`, `Done`, and `Remove` command requirements (the `Root command` and all note requirements are retained).
- `todo-storage`: removes `Markdown file format`, `Storage directory management`, `List discovery`, and `Atomic file writes`.
- `todo-operations`: removes `Add todo`, `Complete todo`, `Remove todo`, and `List todos`.
- `list-management`: removes `Show all lists`, `Create list`, `Rename list`, and `Delete list`.

### Archive Ordering

Because the removals above target requirements introduced by other active changes, this change MUST be archived **after** `go-cli-todo-app`, `list-management`, and `add-configure-command` so those requirements exist in the canonical specs at archive time. The new `notes` capability and the `--dir` / `default_dir` storage-directory resolution are intentionally left intact because notes still use the local `Store`.

## Impact

- Affected commands: `add`, `list`, `done`, `rm`, `lists`, `configure`, plus new `login`, `register`, `logout`, and `undone` (reopen) commands.
- Affected packages: `cmd`, `config`, `internal/store`, and likely new API/client packages for Todex integration.
- Dependencies: add OpenAPI client generation/runtime and standard HTTP/JWT configuration support as needed.
- Runtime behavior: todos and lists require network access and a valid bearer token; local markdown files no longer serve as todo/list storage.
- Out of scope: note commands and note storage remain unchanged for now.
