## MODIFIED Requirements

### Requirement: Note add command
The CLI SHALL provide a `godos note add <title>` command that creates a remote Todex note and opens its body in the user's `$EDITOR` (defaulting to `vi`). The command SHALL accept an optional `--folder <name>` flag that resolves a remote note folder by name.

#### Scenario: Create and edit remote note
- **WHEN** user runs `godos note add "Meeting notes"`
- **THEN** the CLI SHALL create a remote note with title `Meeting notes`
- **AND** the CLI SHALL open the remote note body in `$EDITOR`

#### Scenario: Create note in folder
- **WHEN** user runs `godos note add "Meeting notes" --folder Work`
- **THEN** the CLI SHALL create the remote note in the remote note folder named `Work`

#### Scenario: Missing note folder
- **WHEN** user runs `godos note add "Meeting notes" --folder Missing` and no remote note folder has that name
- **THEN** the CLI SHALL print an error message and exit with a non-zero code

### Requirement: Note show command
The CLI SHALL provide a `godos note show <note-id-prefix>` command that prints the remote note's body to stdout.

#### Scenario: Show existing remote note
- **WHEN** user runs `godos note show abc12345` and the prefix uniquely matches a remote note
- **THEN** the CLI SHALL print the note body to stdout

#### Scenario: Show nonexistent remote note
- **WHEN** user runs `godos note show missing` and the prefix does not match a remote note
- **THEN** the CLI SHALL print an error message and exit with a non-zero code

#### Scenario: Show ambiguous remote note prefix
- **WHEN** user runs `godos note show abc` and the prefix matches multiple remote notes
- **THEN** the CLI SHALL print an ambiguity error and exit with a non-zero code

### Requirement: Note edit command
The CLI SHALL provide a `godos note edit <note-id-prefix>` command that opens an existing remote note body in the user's `$EDITOR` (defaulting to `vi`) and saves changes back to Todex.

#### Scenario: Edit existing remote note
- **WHEN** user runs `godos note edit abc12345` and the prefix uniquely matches a remote note
- **THEN** the CLI SHALL open the note body in `$EDITOR`
- **AND** the CLI SHALL update the remote note body after the editor exits successfully

#### Scenario: Edit nonexistent remote note
- **WHEN** user runs `godos note edit missing` and the prefix does not match a remote note
- **THEN** the CLI SHALL print an error message and exit with a non-zero code

### Requirement: Note remove command
The CLI SHALL provide a `godos note rm <note-id-prefix>` command that soft-deletes a remote note. The CLI SHALL prompt for confirmation before deleting. An optional `--force` flag SHALL skip the confirmation prompt.

#### Scenario: Remove remote note with confirmation
- **WHEN** user runs `godos note rm abc12345` and confirms
- **THEN** the CLI SHALL soft-delete the matching remote note and print a confirmation

#### Scenario: Remove remote note cancelled
- **WHEN** user runs `godos note rm abc12345` and declines confirmation
- **THEN** the CLI SHALL not delete the remote note

#### Scenario: Remove remote note with force flag
- **WHEN** user runs `godos note rm abc12345 --force`
- **THEN** the CLI SHALL soft-delete the matching remote note without prompting

#### Scenario: Remove nonexistent remote note
- **WHEN** user runs `godos note rm missing` and the prefix does not match a remote note
- **THEN** the CLI SHALL print an error message and exit with a non-zero code

### Requirement: Notes list command
The CLI SHALL provide a `godos notes` command that displays remote notes with shortened note ID prefixes and titles. The command SHALL support optional `--folder <name>`, `--query <text>`, `--pinned`, and `--deleted` filters.

#### Scenario: List remote notes
- **WHEN** user runs `godos notes` and remote notes exist
- **THEN** the CLI SHALL display each note's short ID and title

#### Scenario: No remote notes
- **WHEN** user runs `godos notes` and no matching remote notes exist
- **THEN** the CLI SHALL print a message indicating no notes found

#### Scenario: List remote notes with filters
- **WHEN** user runs `godos notes` with folder, query, pinned, or deleted filters
- **THEN** the CLI SHALL display only notes returned by Todex for those filters

## ADDED Requirements

### Requirement: Note restore command
The CLI SHALL provide a `godos note restore <note-id-prefix>` command that restores a soft-deleted remote note.

#### Scenario: Restore remote note
- **WHEN** user runs `godos note restore abc12345` and the prefix uniquely matches a remote note
- **THEN** the CLI SHALL restore the remote note and print a confirmation

### Requirement: Note pin commands
The CLI SHALL provide `godos note pin <note-id-prefix>` and `godos note unpin <note-id-prefix>` commands that update a remote note's pinned state.

#### Scenario: Pin remote note
- **WHEN** user runs `godos note pin abc12345` and the prefix uniquely matches a remote note
- **THEN** the CLI SHALL pin the remote note and print a confirmation

#### Scenario: Unpin remote note
- **WHEN** user runs `godos note unpin abc12345` and the prefix uniquely matches a remote note
- **THEN** the CLI SHALL unpin the remote note and print a confirmation

### Requirement: Goals list command
The CLI SHALL provide a `godos goals` command that displays remote goals with shortened goal ID prefixes, titles, and progress percentages.

#### Scenario: List remote goals
- **WHEN** user runs `godos goals` and remote goals exist
- **THEN** the CLI SHALL display each goal's short ID, title, and progress

#### Scenario: No remote goals
- **WHEN** user runs `godos goals` and no remote goals exist
- **THEN** the CLI SHALL print a message indicating no goals found

### Requirement: Goal add command
The CLI SHALL provide a `godos goal add <title>` command with optional `--description <text>` and `--reason <text>` flags.

#### Scenario: Create remote goal
- **WHEN** user runs `godos goal add "Launch beta" --description "Ship MVP" --reason "Learn faster"`
- **THEN** the CLI SHALL create a remote goal and print its short ID

### Requirement: Goal show command
The CLI SHALL provide a `godos goal show <goal-id-prefix>` command that displays a remote goal's title, description, reason, and progress.

#### Scenario: Show remote goal
- **WHEN** user runs `godos goal show abc12345` and the prefix uniquely matches a remote goal
- **THEN** the CLI SHALL display the remote goal details

### Requirement: Goal edit command
The CLI SHALL provide a `godos goal edit <goal-id-prefix>` command with optional `--title <text>`, `--description <text>`, and `--reason <text>` flags.

#### Scenario: Edit remote goal
- **WHEN** user runs `godos goal edit abc12345 --title "Launch v2"`
- **THEN** the CLI SHALL update the matching remote goal title

### Requirement: Goal remove command
The CLI SHALL provide a `godos goal rm <goal-id-prefix>` command that deletes a remote goal. The CLI SHALL prompt for confirmation before deleting. An optional `--force` flag SHALL skip the confirmation prompt.

#### Scenario: Remove remote goal with confirmation
- **WHEN** user runs `godos goal rm abc12345` and confirms
- **THEN** the CLI SHALL delete the matching remote goal and print a confirmation

#### Scenario: Remove remote goal with force flag
- **WHEN** user runs `godos goal rm abc12345 --force`
- **THEN** the CLI SHALL delete the matching remote goal without prompting

### Requirement: Goal task link commands
The CLI SHALL provide `godos goal link <goal-id-prefix> <task-id-prefix>` and `godos goal unlink <goal-id-prefix> <task-id-prefix>` commands.

#### Scenario: Link task to goal
- **WHEN** user runs `godos goal link abc12345 def67890` and both prefixes are unique
- **THEN** the CLI SHALL link the matching remote task to the matching remote goal

#### Scenario: Unlink task from goal
- **WHEN** user runs `godos goal unlink abc12345 def67890` and both prefixes are unique
- **THEN** the CLI SHALL unlink the matching remote task from the matching remote goal
