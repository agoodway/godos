---
name: godos
description: "Use when the user says '/godos', 'godos', 'todo list', 'manage todos', 'add todo', 'complete todo', 'reopen todo', 'remove todo', 'manage lists', 'godos notes', 'godos note', 'godos goals', 'godos goal', 'Todex API', or wants to install, configure, authenticate, or use the godos CLI."
---

# godos

CLI client for managing Todex-backed todos, lists, notes, and goals.

## Agent Compatibility

This skill is written for Claude Code, Codex, OpenCode, and other AI agents. Use the shell or terminal tool available in the current agent environment to run `godos` commands. Do not rely on agent-specific command tools.

If the user invokes `/godos ...`, treat everything after `/godos` as the `godos` CLI arguments. Example: `/godos note show abc12345` means run or advise `godos note show abc12345`.

## Subcommands

| Subcommand | Purpose |
|------------|---------|
| `install` | Install or build the godos CLI |
| `configure` | Manage API URL and token configuration |
| `login` | Authenticate with a Todex account |
| `register` | Create a Todex account and authenticate |
| `logout` | Clear the saved Todex session |
| `auth` | Inspect authentication state |
| `add` | Add a todo task |
| `list` | Show todos |
| `done` | Mark a todo complete |
| `undone` | Mark a todo incomplete |
| `rm` | Remove a todo |
| `lists` | Manage todo lists |
| `notes` | List remote notes |
| `note` | Manage remote notes |
| `goals` | List remote goals |
| `goal` | Manage remote goals |
| `completion` | Generate shell completion scripts |
| `version` | Print CLI version |
| `help` | Show help |

### `/godos help`

Display a list of available subcommands. Output the following exactly:

```text
/godos subcommands:

  install                         - Install or build the godos CLI
  configure <action>              - Manage API URL and token configuration
  login <email>                   - Authenticate with a Todex account
  register <email>                - Create a Todex account and authenticate
  logout                          - Clear the saved Todex session
  auth status                     - Show the authenticated Todex user
  add <text>                      - Add a todo
  list                            - Show todos
  done <id-prefix>                - Mark a todo complete
  undone <id-prefix>              - Mark a todo incomplete
  rm <id-prefix>                  - Remove a todo
  lists <action>                  - Manage todo lists
  notes [filters]                 - List remote notes
  note <action>                   - Manage remote notes
  goals                           - List remote goals
  goal <action>                   - Manage remote goals
  completion <shell>              - Generate shell completion scripts
  version                         - Print CLI version
  help                            - Show this help message
```

If `/godos` is invoked without a subcommand, show the help output above.

## Dispatch

1. Parse the subcommand and args from the user's invocation. Examples:
   - `/godos install` -> subcommand `install`
   - `/godos configure set api_base_url https://api.example.com` -> subcommand `configure`, action `set`
   - `/godos login user@example.com` -> subcommand `login`
   - `/godos add --list work Ship the feature` -> subcommand `add`
   - `/godos done abc12345` -> subcommand `done`
   - `/godos lists create work` -> subcommand `lists`, action `create`
   - `/godos notes --folder Work --query plan` -> subcommand `notes`
   - `/godos note edit abc12345` -> subcommand `note`, action `edit`
   - `/godos goal add "Launch beta" --description "Ship MVP"` -> subcommand `goal`, action `add`
2. If the subcommand is unknown, list available subcommands and stop.
3. Before running any subcommand except `install`, `help`, and `version`, check if `godos` is installed with the agent's shell tool: `which godos`. If not installed, tell the user to run `/godos install` first and stop.
4. Before running API-backed commands except `install`, `configure`, `login`, `register`, `logout`, `help`, and `version`, ensure configuration exists:
   - `godos configure get api_base_url`
   - Prefer `godos auth status` to check authentication without revealing secrets.
   - Use `godos configure get api_token --show-secret` only when explicitly needed for debugging. Never print this value back to the user.
5. Preserve user-provided flags and quoted text. Do not rewrite command arguments unless needed to make the command syntactically valid.

---

## `/godos install`

Install or build the godos CLI.

1. Check if `godos` is already installed: `which godos`.
2. If already installed, show `godos version` and ask if the user wants to reinstall.
3. If this repository is checked out, build from source:
   ```sh
   go build -o godos .
   ```
4. Otherwise install with Go:
   ```sh
   go install github.com/goodway/godos@latest
   ```
5. Verify installation: `godos version`.
6. Suggest running `/godos configure set api_base_url <url>` and `/godos login <email>`.

---

## `/godos configure`

Manage persistent configuration in `~/.config/godos/config.yaml` (respects `XDG_CONFIG_HOME`). Environment variables override API config values at runtime.

### Actions

| Action | Command |
|--------|---------|
| `set` | `godos configure set <key> <value>` |
| `get` | `godos configure get <key>` |
| `list` | `godos configure list` |
| `delete` | `godos configure delete <key>` |

### Keys

| Key | Purpose | Environment override |
|-----|---------|----------------------|
| `api_base_url` | Todex API base URL | `GODOS_API_BASE_URL` |
| `api_token` | Todex API bearer token | `GODOS_API_TOKEN` |

### Workflow

1. If the user provided an action and args, pass them through directly.
2. If no action is provided, show `godos configure --help` or current config with `godos configure list`.
3. To configure an API URL:
   ```sh
   godos configure set api_base_url <url>
   ```
4. Do not print secrets. `api_token` is masked in `set` and `list`; `get api_token` refuses unless `--show-secret` is used.

---

## `/godos login`, `/godos register`, `/godos logout`, `/godos auth`

Authentication commands use the configured Todex API URL.

```sh
godos login <email>
godos register <email>
godos logout
godos auth status
```

Workflow:

1. Ensure `api_base_url` is configured or `GODOS_API_BASE_URL` is set.
2. For `login` and `register`, run the command and let `godos` prompt for the password. Do not ask the user to reveal the password in chat.
3. Use `godos auth status` to confirm authentication.
4. If auth fails, suggest `godos login <email>` or checking `api_base_url`.

---

## Todos And Lists

Todos and lists are remote Todex resources. Use displayed short ID prefixes, not display positions.

### `/godos add`

```sh
godos add "Write tests"
godos add --list work "Review pull request"
```

Parse all remaining words as todo text unless the user already supplied quoted text or flags. Preserve `--list`.

### `/godos list`

```sh
godos list
godos list --list shopping
godos list --all
```

### `/godos done`, `/godos undone`, `/godos rm`

```sh
godos done <id-prefix>
godos undone <id-prefix>
godos rm <id-prefix>
```

If the user gives a number like `1`, first run `godos list` and ask them to choose the displayed ID prefix. Do not guess a task ID.

### `/godos lists`

| Action | Command |
|--------|---------|
| Show summaries | `godos lists` |
| Create | `godos lists create <name>` |
| Rename | `godos lists rename <old> <new>` |
| Delete | `godos lists delete <name>` |
| Force delete | `godos lists delete <name> --force` |

If deleting without `--force`, let the CLI prompt for confirmation.

---

## Notes

Notes are remote Todex notes. Current command behavior does not manage local markdown note files. Note body creation and editing opens `$EDITOR` and defaults to `vi` if `$EDITOR` is unset.

### `/godos notes`

List remote notes with short IDs and titles.

```sh
godos notes
godos notes --folder Work
godos notes --query "launch plan"
godos notes --pinned
godos notes --deleted
```

Flags may be combined when the CLI supports the combination. Use listed note ID prefixes with `note show`, `note edit`, `note rm`, `note restore`, `note pin`, and `note unpin`.

### `/godos note`

| Action | Command | Notes |
|--------|---------|-------|
| Add | `godos note add <title>` | Opens `$EDITOR` for body content |
| Add to folder | `godos note add <title> --folder <name>` | Resolves a remote note folder by name |
| Show | `godos note show <id-prefix>` | Prints body, or `(empty note)` |
| Edit | `godos note edit <id-prefix>` | Fetches body, opens `$EDITOR`, then updates Todex |
| Remove | `godos note rm <id-prefix>` | Prompts before soft delete |
| Force remove | `godos note rm <id-prefix> --force` | Skips confirmation |
| Restore | `godos note restore <id-prefix>` | Restores a soft-deleted note |
| Pin | `godos note pin <id-prefix>` | Pins a note |
| Unpin | `godos note unpin <id-prefix>` | Unpins a note |

When a note command opens an editor, do not ask the user to paste note content into chat unless they explicitly want that workflow. Let the local editor handle body content.

---

## Goals

Goals are remote Todex goals. Goal progress is display-only and comes from linked tasks.

### `/godos goals`

List remote goals with short IDs, titles, and progress percentages.

```sh
godos goals
```

### `/godos goal`

| Action | Command | Notes |
|--------|---------|-------|
| Add | `godos goal add <title>` | Creates a goal |
| Add with fields | `godos goal add <title> --description <text> --reason <text>` | Optional fields |
| Show | `godos goal show <goal-id-prefix>` | Shows title, description, reason, progress |
| Edit | `godos goal edit <goal-id-prefix> --title <text>` | At least one editable flag required |
| Edit fields | `godos goal edit <goal-id-prefix> --description <text> --reason <text>` | `progress` is not editable |
| Remove | `godos goal rm <goal-id-prefix>` | Prompts before deletion |
| Force remove | `godos goal rm <goal-id-prefix> --force` | Skips confirmation |
| Link task | `godos goal link <goal-id-prefix> <task-id-prefix>` | Resolves both prefixes |
| Unlink task | `godos goal unlink <goal-id-prefix> <task-id-prefix>` | Resolves both prefixes |

If the user asks to change goal progress directly, explain that progress is derived from linked tasks. Use `goal link` or `goal unlink` to affect progress.

---

## `/godos version`

```sh
godos version
```

Use this to confirm installation or report the installed version.

---

## Global Safety

- Do not invent subcommands. If unsure, run `godos --help` or `godos <subcommand> --help`.
- Do not expose `api_token` in chat, logs, or final answers.
- Do not pass passwords as command-line arguments.
- Do not guess todo, note, goal, or task IDs from display order; use the short ID prefixes shown by `godos list`, `godos notes`, or `godos goals`.
- Do not use local markdown note commands for current user-facing note behavior; notes are remote Todex resources.
- Let confirmation prompts run unless the user explicitly supplied `--force` or asked to skip confirmation.
- Preserve user-provided flags and quoted text.

## Common Mistakes

| Mistake | Correct behavior |
|---------|------------------|
| Treating `/godos note add` like local file creation | It creates a remote Todex note and opens `$EDITOR` for body content |
| Using list position `1` as an ID | Run the list command and use the displayed short ID prefix |
| Editing goal progress directly | Link or unlink tasks; progress is display-only |
| Printing `api_token` for debugging | Use `auth status` or inspect locally without revealing the token |
| Assuming only Claude Code supports this skill | Use whichever shell/terminal tool the current agent provides |

$ARGUMENTS
