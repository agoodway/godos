## ADDED Requirements

### Requirement: Add todo
The system SHALL append a new incomplete todo to the end of the specified list. If the list file does not exist, the system SHALL create it.

#### Scenario: Add to existing list
- **WHEN** user adds "Write docs" to a list with 2 existing todos
- **THEN** the list SHALL contain 3 todos with "Write docs" as the last item, marked incomplete

#### Scenario: Add to new list
- **WHEN** user adds "First task" to a list that does not exist
- **THEN** the system SHALL create the list file with a single incomplete todo "First task"

### Requirement: Complete todo
The system SHALL mark a todo as complete by changing `- [ ]` to `- [x]` at the specified 1-based line number. Completing an already-complete todo SHALL be a no-op with a message.

#### Scenario: Complete an incomplete todo
- **WHEN** user completes todo #2 and it is currently `- [ ] Write tests`
- **THEN** the todo SHALL become `- [x] Write tests`

#### Scenario: Complete an already-complete todo
- **WHEN** user completes todo #1 and it is already `- [x] Set up project`
- **THEN** the system SHALL print a message indicating it is already done and make no changes

### Requirement: Remove todo
The system SHALL remove a todo at the specified 1-based line number. Remaining todos SHALL shift to fill the gap (line numbers update).

#### Scenario: Remove middle todo
- **WHEN** a list has 3 todos and user removes todo #2
- **THEN** the list SHALL contain 2 todos; the former #3 becomes #2

### Requirement: List todos
The system SHALL return all todos in a list with their 1-based number, completion status, and text. An empty or nonexistent list SHALL return an empty result.

#### Scenario: List with mixed status
- **WHEN** a list contains `- [x] Done task` and `- [ ] Pending task`
- **THEN** the system SHALL return both with their status and numbers (1: done, 2: pending)

#### Scenario: List nonexistent list
- **WHEN** user lists a list that has no file
- **THEN** the system SHALL return an empty result without error
