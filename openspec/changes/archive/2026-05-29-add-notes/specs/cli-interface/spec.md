## ADDED Requirements

### Requirement: Note add command
The CLI SHALL provide a `godos note add <name>` command that creates a new note and opens it in the user's `$EDITOR` (defaulting to `vi`). The name SHALL be validated before creation.

#### Scenario: Create and edit a note
- **WHEN** user runs `godos note add meeting-notes`
- **THEN** the CLI SHALL create the note file and open it in `$EDITOR`

#### Scenario: Duplicate note name
- **WHEN** user runs `godos note add meeting-notes` and the note already exists
- **THEN** the CLI SHALL print an error message and exit with a non-zero code

#### Scenario: Invalid note name
- **WHEN** user runs `godos note add "invalid name!"`
- **THEN** the CLI SHALL print an error message and exit with a non-zero code

### Requirement: Note show command
The CLI SHALL provide a `godos note show <name>` command that prints the note's contents to stdout.

#### Scenario: Show existing note
- **WHEN** user runs `godos note show meeting-notes` and the note has content
- **THEN** the CLI SHALL print the note contents to stdout

#### Scenario: Show nonexistent note
- **WHEN** user runs `godos note show missing`
- **THEN** the CLI SHALL print an error message and exit with a non-zero code

### Requirement: Note edit command
The CLI SHALL provide a `godos note edit <name>` command that opens an existing note in the user's `$EDITOR` (defaulting to `vi`).

#### Scenario: Edit existing note
- **WHEN** user runs `godos note edit meeting-notes`
- **THEN** the CLI SHALL open the note file in `$EDITOR`

#### Scenario: Edit nonexistent note
- **WHEN** user runs `godos note edit missing`
- **THEN** the CLI SHALL print an error message and exit with a non-zero code

### Requirement: Note remove command
The CLI SHALL provide a `godos note rm <name>` command that deletes a note. The CLI SHALL prompt for confirmation before deleting. An optional `--force` flag SHALL skip the confirmation prompt.

#### Scenario: Remove with confirmation
- **WHEN** user runs `godos note rm meeting-notes` and confirms
- **THEN** the CLI SHALL delete the note and print a confirmation

#### Scenario: Remove cancelled
- **WHEN** user runs `godos note rm meeting-notes` and declines confirmation
- **THEN** the CLI SHALL not delete the note

#### Scenario: Remove with force flag
- **WHEN** user runs `godos note rm meeting-notes --force`
- **THEN** the CLI SHALL delete the note without prompting

#### Scenario: Remove nonexistent note
- **WHEN** user runs `godos note rm missing`
- **THEN** the CLI SHALL print an error message and exit with a non-zero code

### Requirement: Notes list command
The CLI SHALL provide a `godos notes` command that displays all notes with their names.

#### Scenario: List notes
- **WHEN** user runs `godos notes` and two notes exist
- **THEN** the CLI SHALL display each note name, one per line

#### Scenario: No notes
- **WHEN** user runs `godos notes` and no notes exist
- **THEN** the CLI SHALL print a message indicating no notes found
