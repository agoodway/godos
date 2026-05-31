## ADDED Requirements

### Requirement: Todex goal API client
The system SHALL use the configured Todex API base URL and bearer token for all remote goal operations.

#### Scenario: Authenticated goal request
- **WHEN** a goal command contacts Todex
- **THEN** the system SHALL send the configured bearer token using the same authentication behavior as todo and list commands

#### Scenario: Missing goal API configuration
- **WHEN** a goal command runs without a configured Todex base URL or bearer token
- **THEN** the system SHALL return a clear configuration or authentication error

### Requirement: Remote goal listing
The system SHALL list goals from Todex and display each goal with a shortened UUID prefix, title, and progress percentage.

#### Scenario: List goals
- **WHEN** the user lists goals and Todex returns goals
- **THEN** the system SHALL display each goal with its short ID, title, and progress

#### Scenario: Empty goals
- **WHEN** Todex returns no goals
- **THEN** the system SHALL print a message indicating no goals were found

### Requirement: Remote goal creation
The system SHALL create goals in Todex with title, description, and reason fields.

#### Scenario: Create goal
- **WHEN** the user creates a goal with a title, description, and reason
- **THEN** the system SHALL send those fields to Todex and display the created goal's short ID

#### Scenario: Create goal without optional fields
- **WHEN** the user creates a goal with only a title
- **THEN** the system SHALL create the remote goal without setting description or reason

### Requirement: Remote goal retrieval
The system SHALL retrieve a goal from Todex by resolving a user-provided goal ID prefix to one full remote goal UUID.

#### Scenario: Show goal
- **WHEN** the user shows a goal with a unique goal ID prefix
- **THEN** the system SHALL display the remote goal's title, description, reason, and progress

#### Scenario: Missing goal prefix
- **WHEN** no remote goal ID matches the supplied prefix
- **THEN** the system SHALL return a goal not found error

#### Scenario: Ambiguous goal prefix
- **WHEN** multiple remote goal IDs match the supplied prefix
- **THEN** the system SHALL reject the prefix and tell the user to provide more characters

### Requirement: Remote goal update
The system SHALL update goal title, description, and reason fields in Todex by resolved goal ID prefix.

#### Scenario: Update goal fields
- **WHEN** the user updates a goal with a unique goal ID prefix and one or more editable fields
- **THEN** the system SHALL PATCH those fields to Todex

#### Scenario: Goal progress is display-only
- **WHEN** the user updates a goal
- **THEN** the system SHALL NOT send progress as an editable field

### Requirement: Remote goal deletion
The system SHALL delete goals in Todex by resolved goal ID prefix.

#### Scenario: Delete goal with confirmation
- **WHEN** the user deletes a goal and confirms the prompt
- **THEN** the system SHALL call the Todex delete goal endpoint

#### Scenario: Delete goal with force
- **WHEN** the user deletes a goal with the force flag
- **THEN** the system SHALL call the Todex delete goal endpoint without prompting

#### Scenario: Delete goal cancelled
- **WHEN** the user declines the delete confirmation
- **THEN** the system SHALL NOT delete the remote goal

### Requirement: Remote goal task association
The system SHALL link and unlink remote tasks to remote goals by resolving both command arguments from ID prefixes to full UUIDs.

#### Scenario: Link task to goal
- **WHEN** the user links a unique goal ID prefix to a unique task ID prefix
- **THEN** the system SHALL call the Todex link-goal-task endpoint with both full UUIDs

#### Scenario: Unlink task from goal
- **WHEN** the user unlinks a unique task ID prefix from a unique goal ID prefix
- **THEN** the system SHALL call the Todex unlink-goal-task endpoint with both full UUIDs

#### Scenario: Missing linked resource prefix
- **WHEN** either the goal prefix or task prefix does not match a remote resource
- **THEN** the system SHALL return a not found error for the missing resource
