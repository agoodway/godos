## ADDED Requirements

### Requirement: Note storage directory
The system SHALL store notes as individual markdown files in a `notes/` subdirectory within the storage directory. Each note file SHALL be named `<name>.md` where `<name>` follows the same validation rules as list names.

#### Scenario: Notes directory location
- **WHEN** the storage directory is `~/.godos/`
- **THEN** notes SHALL be stored at `~/.godos/notes/<name>.md`

#### Scenario: Auto-create notes directory
- **WHEN** a user creates a note and the `notes/` subdirectory does not exist
- **THEN** the system SHALL create the `notes/` subdirectory before writing the file

### Requirement: Create note
The system SHALL create a new note as an empty `.md` file with the given name. The system SHALL return an error if a note with that name already exists. The name SHALL be validated using the same rules as list names.

#### Scenario: Create new note
- **WHEN** user creates a note named "meeting-notes"
- **THEN** the system SHALL create `notes/meeting-notes.md` in the storage directory

#### Scenario: Create duplicate note
- **WHEN** user creates a note named "meeting-notes" and it already exists
- **THEN** the system SHALL return an error indicating the note already exists

#### Scenario: Invalid note name
- **WHEN** user creates a note with name "../escape"
- **THEN** the system SHALL return an error indicating the name is invalid

### Requirement: Read note
The system SHALL return the contents of a note file as a string. The system SHALL return an error if the note does not exist.

#### Scenario: Read existing note
- **WHEN** user reads note "meeting-notes" which contains text
- **THEN** the system SHALL return the full file contents

#### Scenario: Read nonexistent note
- **WHEN** user reads note "missing" which does not exist
- **THEN** the system SHALL return an error indicating the note was not found

### Requirement: Write note
The system SHALL write content to a note file using atomic writes (temp file + rename), consistent with how list files are written.

#### Scenario: Write note content
- **WHEN** the system writes content to note "meeting-notes"
- **THEN** the content SHALL be written atomically to `notes/meeting-notes.md`

### Requirement: Delete note
The system SHALL remove a note file. The system SHALL return an error if the note does not exist.

#### Scenario: Delete existing note
- **WHEN** user deletes note "meeting-notes"
- **THEN** the file `notes/meeting-notes.md` SHALL be removed

#### Scenario: Delete nonexistent note
- **WHEN** user deletes note "missing" which does not exist
- **THEN** the system SHALL return an error indicating the note was not found

### Requirement: List notes
The system SHALL return all note names by scanning for `.md` files in the `notes/` subdirectory. If the notes directory does not exist, the system SHALL return an empty list.

#### Scenario: List with notes present
- **WHEN** the notes directory contains `ideas.md` and `meeting-notes.md`
- **THEN** the system SHALL return two note names: "ideas" and "meeting-notes"

#### Scenario: List with no notes directory
- **WHEN** the notes directory does not exist
- **THEN** the system SHALL return an empty list without error
