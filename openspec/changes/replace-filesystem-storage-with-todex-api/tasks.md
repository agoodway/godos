## 1. API Client Setup

- [x] 1.1 Vendor the Todex OpenAPI 3.0 contract into the repo (e.g. `api/openapi.json`, copied from `/Users/tbrewer/projects/goodway/todex/openapi.json`) and add a `go:generate` directive so client generation does not depend on the untracked absolute path
- [x] 1.2 Generate and commit Todex client/models using `github.com/oapi-codegen/oapi-codegen` for auth, lists, tasks, and error responses; add the runtime dependency to `go.mod`
- [x] 1.3 Add a client constructor that applies configured base URL and bearer token headers
- [x] 1.4 Add API error normalization for validation, unauthorized, not found, and connection failures

## 2. API Configuration

- [x] 2.1 Extend config support for Todex base URL and bearer token values
- [x] 2.2 Add environment overrides for API base URL and bearer token for tests/automation
- [x] 2.3 Ensure configuration output masks or omits stored bearer tokens
- [x] 2.4 Return clear errors when authenticated commands are missing base URL or token

## 3. Authentication Commands

- [x] 3.1 Implement `godos login <email> <password>` using `POST /api/auth/login`
- [x] 3.2 Implement `godos register <email> <password>` using `POST /api/auth/register`
- [x] 3.3 Persist returned JWT bearer tokens only after successful login/register responses
- [x] 3.4 Add current-user/token validation support using the Todex current-user endpoint
- [x] 3.5 Implement `godos logout` to call `POST /api/auth/logout` when a token exists and clear the local token (clear locally even if the API call fails)
- [x] 3.6 Add command tests for successful auth, failed auth, token persistence, and logout behavior

## 4. Remote Todo/List Service

- [x] 4.1 Add an API-backed service layer for command-oriented list and task operations
- [x] 4.2 Implement remote list discovery, create, rename, delete, and count/summary behavior
- [x] 4.3 Implement list-name to list-UUID resolution with missing and duplicate-name errors
- [x] 4.4 Implement task listing by resolved list name using Todex task APIs
- [x] 4.5 Implement task creation and auto-create missing remote lists for `godos add --list <name>`
- [x] 4.6 Implement task completion and deletion by resolved full task UUID
- [x] 4.7 Implement `godos undone <id-prefix>` task reopen via `POST /api/tasks/{id}/reopen` by resolved full task UUID

## 5. Task ID Prefixes

- [x] 5.1 Display eight-character task UUID prefixes in `godos list` output
- [x] 5.2 Implement task ID prefix resolution across remote tasks
- [x] 5.3 Reject missing task prefixes with task-not-found errors
- [x] 5.4 Reject ambiguous task prefixes with guidance to provide more characters
- [x] 5.5 Reject positional numeric task arguments such as `godos done 3` and `godos rm 3`

## 6. Command Integration

- [x] 6.1 Replace todo/list command use of filesystem store with API-backed service calls
- [x] 6.2 Update `godos add`, `godos list`, `godos done`, and `godos rm` output for remote tasks and ID prefixes
- [x] 6.3 Update `godos lists` create/rename/delete/list behavior for remote lists
- [x] 6.4 Isolate filesystem todo/list code so it is not used as the todo source of truth; retain `internal/store.Store` for note storage (notes still depend on it)
- [x] 6.5 Leave notes commands and note storage behavior unchanged

## 7. Verification

- [x] 7.1 Add fake-server or client-mock tests for remote list/task success paths
- [x] 7.2 Add tests for API auth errors, missing config, missing token, and connection failures
- [x] 7.3 Add tests for task prefix display, unique resolution, missing prefixes, ambiguous prefixes, and numeric positional rejection
- [x] 7.4 Update existing command tests that assumed local markdown todo storage; keep `store_test.go`/`parse_test.go` for the note-backing Store, removing only todo/list-specific assertions
- [x] 7.5 Run `go test ./...` and fix failures
- [x] 7.6 Run project quality checks if available
