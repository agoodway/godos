## ADDED Requirements

### Requirement: Todex note API client
The system SHALL use the configured Todex API base URL and bearer token for all remote note and note folder operations.

#### Scenario: Authenticated note request
- **WHEN** a note command contacts Todex
- **THEN** the system SHALL send the configured bearer token using the same authentication behavior as todo and list commands

#### Scenario: Missing note API configuration
- **WHEN** a note command runs without a configured Todex base URL or bearer token
- **THEN** the system SHALL return a clear configuration or authentication error

### Requirement: Remote note folder resolution
The system SHALL resolve user-facing note folder names to Todex note folder UUIDs before creating or filtering notes.

#### Scenario: Resolve folder by name
- **WHEN** the user supplies `--folder Work` and exactly one remote note folder is named `Work`
- **THEN** the system SHALL use that folder's UUID in the Todex note request

#### Scenario: Missing folder name
- **WHEN** the user supplies `--folder Missing` and no remote note folder has that name
- **THEN** the system SHALL return an error indicating the note folder was not found

#### Scenario: Default folder resolution
- **WHEN** the user creates a note without `--folder`
- **THEN** the system SHALL use the Todex default note folder when one exists

### Requirement: Remote note listing
The system SHALL list notes from Todex with optional folder, query, pinned, and deleted filters.

#### Scenario: List active notes
- **WHEN** the user lists notes without filters
- **THEN** the system SHALL display non-deleted remote notes returned by Todex

#### Scenario: List notes with filters
- **WHEN** the user lists notes with folder, query, pinned, or deleted filters
- **THEN** the system SHALL send the corresponding Todex note list query parameters

#### Scenario: Empty remote notes
- **WHEN** Todex returns no matching notes
- **THEN** the system SHALL print a message indicating no notes were found

### Requirement: Remote note creation
The system SHALL create notes in Todex and store note body content in Todex rather than local files.

#### Scenario: Create note in default folder
- **WHEN** the user creates a note with a title and no folder flag
- **THEN** the system SHALL create the note in the resolved default remote note folder

#### Scenario: Create note in named folder
- **WHEN** the user creates a note with `--folder Work`
- **THEN** the system SHALL create the note in the resolved `Work` remote note folder

### Requirement: Remote note retrieval
The system SHALL retrieve a note from Todex by resolving a user-provided note ID prefix to one full remote note UUID.

#### Scenario: Show note
- **WHEN** the user shows a note with a unique note ID prefix
- **THEN** the system SHALL fetch the full remote note and print its body content

#### Scenario: Missing note prefix
- **WHEN** no remote note ID matches the supplied prefix
- **THEN** the system SHALL return a note not found error

#### Scenario: Ambiguous note prefix
- **WHEN** multiple remote note IDs match the supplied prefix
- **THEN** the system SHALL reject the prefix and tell the user to provide more characters

### Requirement: Remote note editing
The system SHALL edit remote note body content through a temporary editor file and update Todex after the editor exits successfully.

#### Scenario: Edit note body
- **WHEN** the user edits a note with a unique note ID prefix
- **THEN** the system SHALL fetch the remote note body, open it in the user's editor, and PATCH the updated body to Todex

#### Scenario: Editor failure
- **WHEN** the editor exits with an error
- **THEN** the system SHALL return an error and SHALL NOT update the remote note body

### Requirement: Remote note lifecycle actions
The system SHALL support soft delete, restore, pin, and unpin actions for remote notes.

#### Scenario: Soft delete note
- **WHEN** the user removes a note by unique note ID prefix
- **THEN** the system SHALL call the Todex soft-delete note endpoint

#### Scenario: Restore note
- **WHEN** the user restores a note by unique note ID prefix
- **THEN** the system SHALL call the Todex restore note endpoint

#### Scenario: Pin note
- **WHEN** the user pins a note by unique note ID prefix
- **THEN** the system SHALL call the Todex pin note endpoint

#### Scenario: Unpin note
- **WHEN** the user unpins a note by unique note ID prefix
- **THEN** the system SHALL call the Todex unpin note endpoint
