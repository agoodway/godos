# Inspector review-update — replace-filesystem-storage-with-todex-api

Reviewed: 2026-05-30. Two specialists (structural+consistency, codebase-alignment+gaps) audited the change; findings were verified against the codebase and the Todex OpenAPI spec, then auto-patched or resolved with the author.

## Verdict

**Ready (pending archive order)** — all in-change findings are patched and the cross-change collisions have been reconciled (see "Cross-change reconciliation" below). The only remaining requirement is operational: archive this change **after** its three prerequisites.

## Verified facts (grounding)

- `config/config.go` stores a flat `map[string]string` and writes with `0600` perms (`config/config.go:75`) — restrictive token persistence already exists.
- Default `--list` is `todo` in code (`cmd/add.go:27`, `cmd/list.go:66`, `cmd/done.go:36`, `cmd/remove.go:32`) — spec's `todo` default is consistent.
- Positional numeric mutation confirmed (`cmd/done.go`, `cmd/remove.go` use `strconv.Atoi`).
- Todex auth token returned at `data.token` (not top-level) — `POST /api/auth/login` / `register`.
- `GET /api/tasks` supports a `list_id` (uuid) query param — resolves design Open Question 3.
- `POST /api/tasks/{id}/complete` and `.../reopen` both exist; `/api/auth/me` and `/api/me` both exist.

## Remaining findings (unpatched)

### Critical

_None._

### Warning

_Both prior Warnings (cross-change collisions; proposal "Modified" vs ADDED wording) are now resolved — see "Cross-change reconciliation" below._

### Suggestion

3. **Terminology drift** — "JWT bearer token" vs "bearer token" vs "auth token" used interchangeably across delta files. Low risk; standardize on "JWT bearer token" in a later pass.
4. **`godos lists` grammar** — bare `godos lists` (list) vs `godos lists create|rename|delete` (verb group) is not spelled out in design.md. Functionally clear from the delta scenarios.
5. **`/api/auth/me` vs `/api/me`** — Todex exposes both identical current-user endpoints; the change consistently uses `/api/auth/me`, which is fine. No action needed.
6. **Duplicate remote list names** — Todex `List.name` has no documented uniqueness constraint and `POST /api/lists` shows no 409, so duplicates may be possible. Task 4.3 already handles duplicate-name resolution defensively, so this is covered.
7. **`godos add "   "` idempotent/empty + already-complete scenarios** — `specs/remote-todo-storage/spec.md` already hedges the already-complete scenario ("or otherwise preserve idempotent success"); acceptable as written.

## Patches applied

6 findings were auto-patched. 4 findings were patched after author guidance. 0 findings were skipped. Cross-change reconciliation (Warning 1) and the related proposal wording (Warning 2) are intentionally left for an external reconciliation pass.

### Auto-patched

1. **cli-interface delta duplicated 5 requirements** — `specs/cli-interface/spec.md` → removed `Login`, `Register`, `Task list output includes task ID prefixes`, `Done command accepts task ID prefix`, `Remove command accepts task ID prefix` (owned by `api-authentication`, `task-id-prefixes`, `remote-todo-storage`). Kept only the CLI-specific `API configuration command support` requirement.
2. **Open Question 3 already answered** — `design.md` → moved to a new `## Resolved Questions` section; `godos list` resolves the list UUID and calls `GET /api/tasks?list_id=<uuid>` (verified the param exists).
3. **Auth token field unspecified** — `design.md` Decision 4 → documented that login/register return the token at `data.token` (nested under `data`, alongside `data.user`), and that it reuses the existing `0600` config persistence.
4. **Default list name unspecified** — `design.md` Decision 5 → recorded `todo` as the default list name (matches existing CLI flags); noted `is_default` as deferred.
5. **OpenAPI spec at untracked absolute path** — `tasks.md:1.1` → made vendoring explicit (copy into `api/openapi.json` + `go:generate`).
6. **Error normalization gaps** — `specs/todex-api-client/spec.md` → added scenarios for `4xx` validation/conflict (`422`/`409`) and for timeout/malformed-response handling.

### Author-guided patches

1. **Client-generation tool unspecified** — `tasks.md:1.2`, `design.md` Decision 3 → committed to `github.com/oapi-codegen/oapi-codegen` with a `go:generate` directive (author chose oapi-codegen over hand-written / openapi-generator).
2. **logout unhandled (Open Question 2)** — `design.md` (resolved), `tasks.md:3.5`, `specs/api-authentication/spec.md` (new `Logout command` requirement, 3 scenarios), `proposal.md` command list → `godos logout` calls `POST /api/auth/logout` when a token exists and always clears the local token even on API failure (author chose "Add godos logout").
3. **reopen endpoint uncovered** — `specs/remote-todo-storage/spec.md` (new `Remote task reopen` requirement), `tasks.md:4.7`, `proposal.md` command list → added `godos undone <id-prefix>` calling `POST /api/tasks/{id}/reopen` (author chose "Add reopen command").
4. **Filesystem Store / test fate ambiguous** — `tasks.md:6.4`, `tasks.md:7.4` → `internal/store.Store` is retained for note storage; only todo/list paths are isolated, and `store_test.go`/`parse_test.go` are kept with only todo-specific assertions removed (author chose "Keep Store for notes").

## Cross-change reconciliation

This change stacks on three unarchived, code-implemented changes (`go-cli-todo-app`, `list-management`, `add-configure-command`). Because it is a **breaking filesystem→API replacement**, it now expresses supersession explicitly via REMOVED/MODIFIED deltas so the canonical specs stay consistent once everything archives.

| Superseded requirement(s) | Owner change | This change's delta | Replacement |
|---|---|---|---|
| `Add command`, `List command`, `Done command`, `Remove command` | go-cli-todo-app (`cli-interface`) | `cli-interface` REMOVED | `remote-todo-storage` + `task-id-prefixes` |
| `Markdown file format`, `Storage directory management`, `List discovery`, `Atomic file writes` | go-cli-todo-app (`todo-storage`) | `todo-storage` REMOVED | `remote-todo-storage` (todo data in Todex) |
| `Add todo`, `Complete todo`, `Remove todo`, `List todos` | go-cli-todo-app (`todo-operations`) | `todo-operations` REMOVED | `remote-todo-storage` + `task-id-prefixes` |
| `Show all lists`, `Create list`, `Rename list`, `Delete list` | list-management (`list-management`) | `list-management` REMOVED | `remote-todo-storage` remote list ops |
| `List all configuration values` | add-configure-command (`configuration`) | `configuration` MODIFIED | masks the stored JWT |

Retained on purpose: `Root command` (`--dir`) and all `notes` requirements — notes still use the local `Store` (`internal/store/store.go:127` `NotesDir`, resolved via `cmd/root.go:29`). The `Lists command group` requirement (list-management) also stays valid; only its filesystem backend changed. No literal duplicate `### Requirement:` names remain across changes.

Two follow-through items were handled while reconciling:
- **Count display preserved** — `remote-todo-storage` → "Remote list discovery" now shows completed/total counts (matching the old `godos lists`), since the Todex `List` schema has no count field and counts must be derived per list.
- **Archive order recorded** — `proposal.md` and `design.md` now state this change MUST archive after `go-cli-todo-app`, `list-management`, and `add-configure-command`.

## Summary

| Bucket | Count |
|--------|-------|
| Auto-patched | 6 |
| Author-guided | 4 |
| Cross-change reconciliation deltas added | 5 (4 REMOVED + 1 MODIFIED) |
| Remaining (Warning) | 0 |
| Remaining (Suggestion) | 5 |

The change is internally consistent and the cross-change collisions are reconciled. Remaining items are low-risk Suggestions only. **Operational requirement:** archive after the three prerequisite changes.
