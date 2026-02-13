## ADDED Requirements

### Requirement: Root command with global flags
The `godos` binary SHALL provide a root command with a `--dir` flag to override the default storage directory (`~/.godos/`). Running `godos` with no subcommand SHALL display help text.

#### Scenario: Default help output
- **WHEN** user runs `godos` with no arguments
- **THEN** the CLI SHALL display available subcommands and global flags

#### Scenario: Custom directory flag
- **WHEN** user runs `godos --dir /tmp/todos list`
- **THEN** the CLI SHALL use `/tmp/todos` as the storage directory instead of `~/.godos/`

### Requirement: Add command
The CLI SHALL provide a `godos add <text>` command that adds a new todo to a list. An optional `--list` flag specifies the target list (defaults to `todo`).

#### Scenario: Add todo to default list
- **WHEN** user runs `godos add "Buy groceries"`
- **THEN** the CLI SHALL append `- [ ] Buy groceries` to `~/.godos/todo.md` and print a confirmation

#### Scenario: Add todo to named list
- **WHEN** user runs `godos add "Fix login bug" --list work`
- **THEN** the CLI SHALL append `- [ ] Fix login bug` to `~/.godos/work.md`

### Requirement: List command
The CLI SHALL provide a `godos list` command that displays all todos in a list with their line numbers and completion status. An optional `--list` flag specifies which list (defaults to `todo`). An optional `--all` flag shows all lists.

#### Scenario: List todos with numbers
- **WHEN** user runs `godos list` and the default list contains 3 todos
- **THEN** the CLI SHALL display each todo with a 1-based number, e.g. `1. [ ] Buy groceries`

#### Scenario: List all lists
- **WHEN** user runs `godos list --all`
- **THEN** the CLI SHALL display todos from every `.md` file in the storage directory, grouped by list name

### Requirement: Done command
The CLI SHALL provide a `godos done <number>` command that marks a todo as complete by its line number. An optional `--list` flag specifies the list.

#### Scenario: Mark todo as done
- **WHEN** user runs `godos done 2` and todo #2 is `- [ ] Write tests`
- **THEN** the CLI SHALL change it to `- [x] Write tests` and print a confirmation

#### Scenario: Invalid number
- **WHEN** user runs `godos done 99` and the list has fewer than 99 todos
- **THEN** the CLI SHALL print an error message and exit with a non-zero code

### Requirement: Remove command
The CLI SHALL provide a `godos rm <number>` command that removes a todo by its line number. An optional `--list` flag specifies the list.

#### Scenario: Remove a todo
- **WHEN** user runs `godos rm 1`
- **THEN** the CLI SHALL remove the first todo from the list and print a confirmation

#### Scenario: Invalid number on remove
- **WHEN** user runs `godos rm 0`
- **THEN** the CLI SHALL print an error message and exit with a non-zero code
