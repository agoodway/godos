---
name: godos
description: "Use when the user says '/godos', 'godos', 'todo list', 'manage todos', 'add todo', 'complete todo', 'reopen todo', 'remove todo', 'manage lists', 'Todex API', or wants to install, configure, authenticate, or use the godos CLI."
---

# godos

CLI client for managing Todex-backed todo lists from the command line.

## Subcommands

| Subcommand | Purpose |
|------------|---------|
| `install` | Install or build the godos CLI |
| `configure` | Manage API URL and token configuration |
| `login` | Authenticate with a Todex account |
| `register` | Create a Todex account and authenticate |
| `logout` | Clear the saved Todex session |
| `auth` | Inspect authentication state |
| `add` | Add a todo |
| `list` | Show todos |
| `done` | Mark a todo complete |
| `undone` | Mark a todo incomplete |
| `rm` | Remove a todo |
| `lists` | Manage todo lists |
| `version` | Print CLI version |

### `/godos help`

Display a list of all available subcommands. Output the following exactly:

```text
/godos subcommands:

  install                  - Install or build the godos CLI
  configure <action>       - Manage API URL and token configuration
  login <email>            - Authenticate with a Todex account
  register <email>         - Create a Todex account and authenticate
  logout                   - Clear the saved Todex session
  auth status              - Show the authenticated Todex user
  add <text>               - Add a todo
  list                     - Show todos
  done <id-prefix>         - Mark a todo complete
  undone <id-prefix>       - Mark a todo incomplete
  rm <id-prefix>           - Remove a todo
  lists <action>           - Manage todo lists
  version                  - Print CLI version
  help                     - Show this help message
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
   - `/godos help` -> show help
2. If the subcommand is unknown, list available subcommands and stop.
3. Before running any subcommand except `install`, `help`, and `version`, check if `godos` is installed (`which godos`). If not, tell the user to run `/godos install` first and stop.
4. Before running API-backed commands except `install`, `configure`, `login`, `register`, `logout`, `help`, and `version`, ensure configuration exists:
   - `godos configure get api_base_url`
   - `godos configure get api_token --show-secret` only when authentication is required and the user explicitly needs to verify the token. Never print this value back to the user.
   - Prefer `godos auth status` to check authentication without revealing secrets.
5. Follow the matching workflow below.

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
6. Suggest running `/godos configure` and `/godos login <email>`.

---

## `/godos configure`

Manage persistent configuration in `~/.config/godos/config.yaml` (respects `XDG_CONFIG_HOME`).

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
| `api_token` | Todex API token | `GODOS_API_TOKEN` |

### Workflow

1. If the user provided an action and args, pass them through directly.
2. If no action is provided, show current config with `godos configure list`.
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

## `/godos add`

Add a todo to the default `todo` list or a named list.

```sh
godos add "Write tests"
godos add --list work "Review pull request"
```

Parse all remaining words as the todo text unless the user already supplied quoted text or flags. Preserve user-provided flags.

---

## `/godos list`

Show todos.

```sh
godos list
godos list --list shopping
godos list --all
```

Notes:
- Todo IDs are short ID prefixes shown by the API-backed CLI.
- Use those prefixes with `done`, `undone`, and `rm`.

---

## `/godos done`, `/godos undone`, `/godos rm`

Operate on todos by ID prefix, not by list position.

```sh
godos done <id-prefix>
godos undone <id-prefix>
godos rm <id-prefix>
```

If the user gives a number like `1`, first run `godos list` and ask them to choose the displayed ID prefix. Do not guess a task ID.

---

## `/godos lists`

Manage todo lists.

| Action | Command |
|--------|---------|
| Show summaries | `godos lists` |
| Create | `godos lists create <name>` |
| Rename | `godos lists rename <old> <new>` |
| Delete | `godos lists delete <name>` |
| Force delete | `godos lists delete <name> --force` |

If deleting without `--force`, let the CLI prompt for confirmation.

---

## `/godos version`

```sh
godos version
```

Use this to confirm installation or report the installed version.

---

## Global Safety

- Do not invent subcommands. If unsure, run `godos --help` or the relevant `godos <subcommand> --help`.
- Do not expose `api_token` in chat or logs.
- Do not pass passwords as command-line arguments.
- Do not guess task IDs from display order; use the short ID prefixes shown by `godos list`.
- Preserve user-provided flags and quoted text.

$ARGUMENTS
