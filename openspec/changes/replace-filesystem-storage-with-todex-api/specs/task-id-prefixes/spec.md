## ADDED Requirements

### Requirement: Short task ID display
The CLI SHALL display a shortened UUID prefix for each task in task list output.

#### Scenario: List output includes short ID
- **WHEN** the user runs `godos list` and Todex returns task ID `9f3a2c1e-1b9a-4f52-b4b7-123456789abc`
- **THEN** the CLI SHALL display `9f3a2c1e` as the task identifier in the list output

#### Scenario: Full task ID not required in normal list output
- **WHEN** the user runs `godos list`
- **THEN** the CLI SHALL NOT require the user to visually parse the full UUID to use common mutation commands

### Requirement: Unique task prefix resolution
The system SHALL resolve user-provided task ID prefixes to full Todex task UUIDs before mutating tasks.

#### Scenario: Prefix resolves to one task
- **WHEN** the user runs `godos done 9f3a2c1e` and exactly one task ID starts with `9f3a2c1e`
- **THEN** the CLI SHALL mutate that task using its full UUID

#### Scenario: Prefix has no matches
- **WHEN** the user runs `godos done deadbeef` and no task ID starts with `deadbeef`
- **THEN** the CLI SHALL exit non-zero with a task-not-found error

#### Scenario: Prefix is ambiguous
- **WHEN** the user runs `godos done 9f3a` and multiple task IDs start with `9f3a`
- **THEN** the CLI SHALL exit non-zero and instruct the user to provide more ID characters

### Requirement: Positional task mutation rejected
The CLI SHALL reject positional task mutation arguments for commands that mutate individual tasks.

#### Scenario: Done with numeric position
- **WHEN** the user runs `godos done 3`
- **THEN** the CLI SHALL reject the argument as an invalid task ID prefix and SHALL NOT complete the third visible task

#### Scenario: Remove with numeric position
- **WHEN** the user runs `godos rm 3`
- **THEN** the CLI SHALL reject the argument as an invalid task ID prefix and SHALL NOT remove the third visible task
