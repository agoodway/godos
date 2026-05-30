## Context

`godos` is currently a local-first CLI that stores todo lists as markdown files under a configurable directory. The command layer calls `internal/store.Store` directly, and `Store` combines local filesystem safety, markdown parsing/rendering, and todo/list operations.

The Todex API exposes the remote model that should replace filesystem todo storage:

- Auth endpoints: `POST /api/auth/register`, `POST /api/auth/login`, `GET /api/auth/me`, `POST /api/auth/logout`
- List endpoints: `GET /api/lists`, `POST /api/lists`, `GET/PATCH/DELETE /api/lists/{id}`
- Task endpoints: `GET /api/tasks`, `POST /api/tasks`, `GET/PATCH/DELETE /api/tasks/{id}`, `POST /api/tasks/{id}/complete`, `POST /api/tasks/{id}/reopen`

The API uses UUIDs for lists and tasks. The existing CLI uses list names and positional todo numbers. This change keeps list names as user-facing selectors but changes task mutations to use stable task ID prefixes.

## Goals / Non-Goals

**Goals:**

- Make todo/list commands API-only for Todex.
- Add `godos register` and `godos login` to obtain and persist JWT bearer tokens.
- Store API configuration locally, including base URL and bearer token.
- Generate or otherwise maintain a typed Todex API client from `/Users/tbrewer/projects/goodway/todex/openapi.json`.
- Preserve familiar list-oriented command UX where possible.
- Auto-create a remote list when `add --list <name>` targets a missing list.
- Display shortened task UUID prefixes and accept unique prefixes for task mutations.

**Non-Goals:**

- Offline/local fallback for todo/list commands.
- Migrating existing markdown todo files into Todex.
- Sync/conflict handling between local files and the API.
- Notes API support; existing note behavior remains out of scope.
- Changing the Todex API contract.

## Decisions

### 1. API-only todo/list backend

Todo and list operations will call Todex over HTTP instead of reading or writing markdown list files. The old filesystem-backed todo/list storage should not remain as a selectable todo backend.

Rationale: The user explicitly chose API-only replacement. Keeping dual backends would add branching behavior and migration complexity that is not needed for this phase.

Alternative considered: Keep filesystem as an offline backend. Rejected because it would require sync semantics and conflict resolution, which are out of scope.

### 2. Keep command layer behind an operation boundary

The implementation should avoid leaking generated OpenAPI types throughout `cmd/*`. Introduce a small API-backed service/client layer that exposes command-oriented operations such as list lookup by name, add task, complete task by prefix, and delete task by prefix.

Rationale: Existing commands are expressed in CLI concepts, while the Todex API is expressed in UUID resources. A thin adapter keeps mapping logic centralized and makes tests easier.

Alternative considered: Call generated client methods directly from each command. Rejected because list-name resolution, auth headers, prefix resolution, and error normalization would be duplicated.

### 3. Generate a Todex API client from OpenAPI 3.0

Use the Todex OpenAPI 3.0 JSON as the source of truth for request/response models and client operations. Generated code should be committed and regenerated when the spec changes.

Rationale: The API contract already exists, and generated models reduce drift for schemas such as `List`, `Task`, `AuthResponse`, `ListRequest`, and `TaskRequest`.

Tooling: use `github.com/oapi-codegen/oapi-codegen` (idiomatic Go, single runtime dependency) driven by a `go:generate` directive against the vendored `api/openapi.json`.

Alternative considered: Hand-write HTTP request structs. This is viable for a small API but increases contract drift risk. Also considered the Java-based `openapi-generator`, rejected to avoid a non-Go toolchain dependency.

### 4. Persist base URL and JWT token in local config

`godos login` and `godos register` will store the returned token. API commands will read the token and send it as `Authorization: Bearer <token>`. Base URL should also be configurable, with an environment variable override suitable for tests and local development.

Rationale: A CLI needs non-interactive repeated use after login. Persisting the JWT avoids requiring credentials for every command.

Alternative considered: Token-only environment variable. Useful as an override, but insufficient as the primary UX.

Response shape note: `POST /api/auth/login` and `POST /api/auth/register` return the token nested as `data.token` (with `data.user`), not as a top-level field. The client/service layer SHALL unwrap `data.token` before persisting it. The `godos` config already stores values as a flat `map[string]string` written with `0600` permissions (`config/config.go`), so the bearer token reuses that restrictive-permission persistence.

### 5. List names remain user-facing; list UUIDs are internal

Commands that accept `--list <name>` will resolve the remote list by name. `godos add --list <name>` will create the list when no remote list with that name exists. Other list-specific commands will error when the named list is missing unless their purpose is creation.

Rationale: The existing CLI is name-oriented and this keeps the common workflow readable. The API UUID remains an implementation detail except when explicitly managing task IDs.

Default list: the current CLI defaults the `--list` flag to `todo` (`cmd/add.go`, `cmd/list.go`, `cmd/done.go`, `cmd/remove.go`). The remote backend SHALL keep `todo` as the default list name resolved by name. Whether to instead honor a Todex `is_default` list flag is deferred; defaulting to the `todo` name preserves existing behavior.

Alternative considered: Require list UUIDs in all commands. Rejected as unnecessarily hostile for a CLI.

### 6. Task mutation commands use task ID prefixes

`godos list` will display shortened task UUID prefixes. `godos done <id-prefix>` and `godos rm <id-prefix>` will resolve prefixes against the authenticated user's tasks, require exactly one match, and then call the API with the full UUID.

Rationale: Positional numbers become unsafe with a remote backend because task ordering can change between `list` and mutation commands. Unique prefixes keep the CLI ergonomic while preserving stable identity.

Alternative considered: Keep `done 3` by fetching and sorting current tasks. Rejected because it can mutate the wrong task after remote changes.

### 7. Default display prefix length is eight characters

Task listings should show the first eight UUID characters by default. The resolver should accept any prefix length and reject ambiguous matches.

Rationale: Eight characters are readable and usually enough for small personal task sets. Ambiguity checks preserve correctness.

Alternative considered: Always show full UUIDs. Correct but noisy.

### 8. Supersede filesystem capabilities via REMOVED deltas, ordered after prerequisites

This change introduces new remote capabilities (`todex-api-client`, `api-authentication`, `remote-todo-storage`, `task-id-prefixes`) rather than rewriting the filesystem-era capabilities in place. To keep the canonical specs consistent after archiving, it adds `## REMOVED Requirements` deltas for the filesystem todo requirements owned by the not-yet-archived `go-cli-todo-app` and `list-management` changes (`todo-storage`, `todo-operations`, `list-management`, and the `cli-interface` Add/List/Done/Remove commands), plus a `## MODIFIED Requirements` delta for `configuration` (`List all configuration values` masks the token).

Supersession mapping:
- `todo-storage` (markdown format, dir mgmt, list discovery, atomic writes) → removed; todo data lives in Todex.
- `todo-operations` (add/complete/remove/list todo) → `remote-todo-storage` + `task-id-prefixes`.
- `list-management` (show/create/rename/delete list) → `remote-todo-storage` remote list operations (count display preserved).
- `cli-interface` Add/List/Done/Remove → `remote-todo-storage` + `task-id-prefixes`; `Root command` and note requirements retained.
- `configuration` `List all configuration values` → modified to mask the stored JWT.

Rationale: a wholesale "remove old, add new" reads more clearly for a breaking replacement than threading remote semantics through each legacy requirement, and the new capability names already fully specify the remote behavior.

Archive ordering: this change MUST archive after `go-cli-todo-app`, `list-management`, and `add-configure-command` so the REMOVED/MODIFIED targets exist in canonical. Notes (`Store`, `--dir`, `default_dir`) are untouched because notes still use local storage.

## Risks / Trade-offs

- [Network dependency makes todo/list commands fail offline] -> Return clear API connection/auth errors; no offline fallback is planned.
- [Stored JWT can be exposed from local config] -> Store config with restrictive file permissions where possible and mask token values in `configure list` or auth-related output.
- [OpenAPI spec path is outside this repo] -> Copy or reference the spec in a reproducible generation setup so future `go generate ./...` does not depend on an untracked absolute path.
- [Task prefix ambiguity] -> Require exactly one matching task; tell the user to provide more characters when ambiguous.
- [List-name duplicates from the API] -> Treat duplicate matching list names as an error unless the API guarantees uniqueness.
- [Existing tests assume filesystem semantics] -> Add API/service tests using fake HTTP servers or generated client mocks; update command tests to target API-backed behavior.

## Migration Plan

1. Add generated Todex API client and API configuration support.
2. Add auth commands and token persistence.
3. Introduce API-backed todo/list operation layer.
4. Update commands to use API-backed operations and task ID prefixes.
5. Remove or isolate filesystem todo/list storage paths from command behavior.
6. Keep note code unchanged and out of the API-backed todo/list path.

Rollback is code-level only: revert this change to restore filesystem-backed todo/list behavior. No data migration is planned.

## Open Questions

- Should base URL default to a specific production Todex URL, or require explicit configuration before login?

## Resolved Questions

- **`godos list` task query strategy**: Resolved. `GET /api/tasks` supports a `list_id` (UUID) query parameter, so `godos list` SHALL resolve the list name to its UUID and call `GET /api/tasks?list_id=<uuid>` rather than fetching all tasks and filtering client-side.
- **Logout behavior**: Resolved. `godos logout` SHALL call `POST /api/auth/logout` when a stored token exists and then clear the local token. If the API call fails, the local token SHALL still be cleared so the user is not stuck authenticated locally.
