## ADDED Requirements

### Requirement: Markdown file format
Each todo list SHALL be stored as a markdown file where each todo is a line matching the pattern `- [ ] <text>` (incomplete) or `- [x] <text>` (complete). Non-todo lines in the file SHALL be preserved during read/write operations.

#### Scenario: Parse incomplete todo
- **WHEN** the storage layer reads a line `- [ ] Buy milk`
- **THEN** it SHALL parse it as an incomplete todo with text "Buy milk"

#### Scenario: Parse complete todo
- **WHEN** the storage layer reads a line `- [x] Buy milk`
- **THEN** it SHALL parse it as a complete todo with text "Buy milk"

#### Scenario: Preserve non-todo lines
- **WHEN** a markdown file contains headings, blank lines, or other content between todo lines
- **THEN** the storage layer SHALL preserve those lines unchanged when writing back

### Requirement: Storage directory management
The storage layer SHALL create the storage directory if it does not exist on first write. The storage layer SHALL NOT create the directory on read-only operations (e.g., listing an empty directory).

#### Scenario: Auto-create directory on add
- **WHEN** user adds a todo and `~/.godos/` does not exist
- **THEN** the storage layer SHALL create `~/.godos/` and write the todo file

#### Scenario: Empty directory on list
- **WHEN** user runs list and the storage directory does not exist
- **THEN** the CLI SHALL display an empty list without errors

### Requirement: List discovery
The storage layer SHALL discover all lists by scanning for `*.md` files in the storage directory. The list name SHALL be the filename without the `.md` extension.

#### Scenario: Discover multiple lists
- **WHEN** the storage directory contains `todo.md`, `work.md`, and `shopping.md`
- **THEN** the storage layer SHALL report three lists: `todo`, `work`, and `shopping`

### Requirement: Atomic file writes
The storage layer SHALL write to a temporary file and rename it to the target path to prevent data loss from interrupted writes.

#### Scenario: Write atomicity
- **WHEN** the storage layer writes an updated todo list
- **THEN** it SHALL write to a temporary file in the same directory and rename it to the target filename
