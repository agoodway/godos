## ADDED Requirements

### Requirement: Show all lists
The system SHALL display all available lists with their name and todo count (total and completed).

#### Scenario: Multiple lists exist
- **WHEN** user runs `godos lists` and the storage directory contains `todo.md` (3 todos, 1 done) and `work.md` (5 todos, 0 done)
- **THEN** the CLI SHALL display each list with its counts, e.g. `todo  (1/3 done)` and `work  (0/5 done)`

#### Scenario: No lists exist
- **WHEN** user runs `godos lists` and the storage directory is empty or does not exist
- **THEN** the CLI SHALL display a message indicating no lists found

### Requirement: Create list
The system SHALL create an empty list file when the user runs `godos lists create <name>`. The list name MUST be valid for use as a filename (alphanumeric, hyphens, underscores).

#### Scenario: Create new list
- **WHEN** user runs `godos lists create shopping`
- **THEN** the system SHALL create `shopping.md` in the storage directory and print a confirmation

#### Scenario: Create duplicate list
- **WHEN** user runs `godos lists create work` and `work.md` already exists
- **THEN** the CLI SHALL print an error indicating the list already exists and exit with a non-zero code

#### Scenario: Invalid list name
- **WHEN** user runs `godos lists create "my list!"` (contains spaces or special characters)
- **THEN** the CLI SHALL print an error indicating the name is invalid and exit with a non-zero code

### Requirement: Rename list
The system SHALL rename a list by renaming its markdown file when the user runs `godos lists rename <old> <new>`.

#### Scenario: Rename existing list
- **WHEN** user runs `godos lists rename shopping groceries` and `shopping.md` exists
- **THEN** the system SHALL rename `shopping.md` to `groceries.md` and print a confirmation

#### Scenario: Rename nonexistent list
- **WHEN** user runs `godos lists rename foo bar` and `foo.md` does not exist
- **THEN** the CLI SHALL print an error indicating the source list does not exist

#### Scenario: Rename to existing name
- **WHEN** user runs `godos lists rename shopping work` and `work.md` already exists
- **THEN** the CLI SHALL print an error indicating the target name is already taken

### Requirement: Delete list
The system SHALL delete a list and all its todos when the user runs `godos lists delete <name>`. The command SHALL prompt for confirmation unless `--force` is provided.

#### Scenario: Delete with confirmation
- **WHEN** user runs `godos lists delete shopping` and `shopping.md` has 4 todos
- **THEN** the CLI SHALL prompt "Delete list 'shopping' with 4 todos? [y/N]" and delete only if user confirms

#### Scenario: Delete with force flag
- **WHEN** user runs `godos lists delete shopping --force`
- **THEN** the system SHALL delete `shopping.md` without prompting

#### Scenario: Delete nonexistent list
- **WHEN** user runs `godos lists delete foo` and `foo.md` does not exist
- **THEN** the CLI SHALL print an error indicating the list does not exist

#### Scenario: Delete default list
- **WHEN** user runs `godos lists delete todo`
- **THEN** the system SHALL allow deleting the default list (no special protection)
