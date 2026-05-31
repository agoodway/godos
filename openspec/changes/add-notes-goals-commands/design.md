## Context

`godos` has already moved todo and list behavior to the Todex REST API behind `internal/todex.Service`, with authentication and base URL configuration handled by the existing config package. Notes are the remaining local-only concept: canonical `notes` specs still describe markdown files under a local storage directory, and `internal/store.Store` contains note file helpers.

Todex now exposes protected REST APIs for note folders, notes, and goals in `/Users/tbrewer/projects/goodway/todex/openapi.json`:

- Note folders: `GET/POST /api/note-folders`, `GET/PATCH/DELETE /api/note-folders/{id}`
- Notes: `GET/POST /api/notes`, `GET/PATCH/DELETE /api/notes/{id}`, plus `pin`, `unpin`, `restore`, and permanent deletion endpoints
- Goals: `GET/POST /api/goals`, `GET/PATCH/DELETE /api/goals/{id}`, plus task link/unlink endpoints

The current `godos/api/openapi.json` is stale and must be refreshed before generated Go code can expose those operations.

## Goals / Non-Goals

**Goals:**

- Make note commands API-only and stop using local note files as command behavior.
- Add command-oriented note operations to `internal/todex.Service` instead of exposing generated OpenAPI types throughout `cmd`.
- Add goal commands for CRUD and task linking/unlinking.
- Reuse the existing Todex base URL, bearer token, HTTP safety, response-size, and error-normalization behavior.
- Use shortened UUID prefixes for notes and goals, matching the task command pattern.
- Keep folder names user-facing for note workflows while keeping folder UUIDs internal.

**Non-Goals:**

- Migrating existing local note files into Todex.
- Offline note or goal fallback.
- WebSocket/realtime client support.
- Full note folder management commands beyond folder-name lookup for listing and creation.
- Setting goal progress directly; progress remains derived and display-only.

## Decisions

### 1. Refresh OpenAPI first, then regenerate the client

The implementation will copy the updated Todex `openapi.json` into `godos/api/openapi.json` and run the existing `go generate` setup for `internal/todexapi`.

Rationale: The project already committed generated OpenAPI code for Todex. Continuing that pattern keeps request and response models aligned with the backend contract and avoids hand-written drift.

Alternative considered: hand-write HTTP requests for notes/goals. Rejected because the API now has enough schemas and operations that duplication would make drift likely.

### 2. Keep `cmd` behind command-oriented service methods

`cmd` should call methods such as `ListNotes`, `CreateNote`, `ShowNote`, `UpdateNote`, `DeleteNote`, `RestoreNote`, `PinNote`, `ListGoals`, `CreateGoal`, `UpdateGoal`, `DeleteGoal`, `LinkGoalTask`, and `UnlinkGoalTask`. The generated OpenAPI types stay inside `internal/todex` except where tests need fake server JSON.

Rationale: Existing todo/list commands already use a small service boundary. Notes and goals need the same central place for prefix resolution, folder-name lookup, response mapping, and error normalization.

Alternative considered: direct generated client calls from each command. Rejected because it would duplicate prefix and lookup logic across commands.

### 3. Notes use ID prefixes, folders use names

Notes are addressed by UUID prefixes in command arguments because note titles are mutable and not guaranteed unique. Note folders remain name-oriented in the CLI: `--folder <name>` resolves to a folder UUID for note list/create requests. If `--folder` is omitted for create, the service resolves the API-provided default folder, falling back to a folder named `Notes` if necessary.

Rationale: This keeps note mutation commands safe while preserving a readable folder UX. It mirrors task prefixes and list-name resolution already used by todos.

Alternative considered: address notes by title. Rejected because duplicate or changed titles can mutate the wrong note.

### 4. Remote note editing uses a temporary file bridge

`godos note edit <note-id-prefix>` will fetch the remote note, write its body to a temporary markdown file, open `$EDITOR` (default `vi`), read the file after the editor exits, and PATCH the remote note body. `godos note add <title>` creates the remote note first, then can use the same editor flow to populate body content when implemented by the command layer.

Rationale: The existing note UX is editor-centered, while the remote API stores note body text. A temp-file bridge preserves CLI ergonomics without reintroducing local persistence.

Alternative considered: require `--body` or stdin-only editing. Rejected as less consistent with the existing note workflow.

### 5. Goal link/unlink resolves both prefixes client-side

Goal commands resolve goal prefixes by listing remote goals. Link/unlink commands resolve the goal prefix through remote goals and task prefix through the existing remote task prefix behavior, then call the Todex goal association endpoints with full UUIDs.

Rationale: Todex link/unlink endpoints require full UUIDs. Prefixes keep the CLI ergonomic while preserving stable identity and ambiguity checks.

Alternative considered: require full UUIDs for goals. Rejected as too noisy for normal CLI usage.

### 6. Local note store becomes legacy-only

Command behavior will stop calling `internal/store.Store` for notes. The implementation can either remove note-specific store methods/tests if no remaining code depends on them or leave them isolated as unused legacy code if removing them would create avoidable churn.

Rationale: The user chose API-backed-only notes. Keeping local notes in command paths would contradict that product decision.

Alternative considered: keep local notes and add separate remote commands. Rejected during discovery because it creates confusing dual storage semantics.

## Risks / Trade-offs

- [Existing local notes become invisible to `godos`] -> Document this as a breaking change; no migration is planned in this change.
- [Prefix ambiguity for notes/goals] -> Require exactly one match and tell users to provide more characters when ambiguous.
- [Default note folder lookup can be inconsistent if the API has duplicate or missing defaults] -> Prefer `is_default`, fall back to name `Notes`, and return a clear error if no usable folder exists.
- [Editor flow can fail after remote note creation] -> Surface editor/read/write errors clearly; do not hide that a note may have been created with an empty body.
- [Generated client changes may affect existing task/list code] -> Regenerate in one step and run the existing command/service test suite against fake HTTP servers.
- [Service file growth] -> Keep mapping helpers small and extract note/goal-specific conversion or prefix helpers only if the file becomes hard to reason about.

## Migration Plan

1. Refresh `api/openapi.json` from Todex and regenerate `internal/todexapi`.
2. Add note and goal service models, conversion helpers, prefix resolution, and command-oriented service methods.
3. Add/replace note commands so they use `getAPIService(true)` and Todex notes instead of `getStore()`.
4. Add goal commands and task association commands.
5. Update command tests and fake remote server support for notes/goals.
6. Remove or isolate local note command/storage behavior from active paths.

Rollback is code-level only: revert this change to restore local note behavior. No local-to-remote migration or remote-to-local rollback is planned.

## Open Questions

- Should `godos note add` always open `$EDITOR` after remote creation, or should there be a `--no-edit` flag for scripts?
