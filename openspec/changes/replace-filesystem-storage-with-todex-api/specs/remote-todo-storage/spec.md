## ADDED Requirements

### Requirement: Remote list discovery
The system SHALL discover todo lists from Todex using the authenticated list API instead of scanning local markdown files. To preserve the existing `godos lists` workflow, the CLI SHALL display each list's todo count (completed/total) alongside its name.

#### Scenario: List remote lists
- **WHEN** the user runs `godos lists`
- **THEN** the CLI SHALL fetch lists from Todex and display each remote list name with its completed/total task count (for example `work  (0/5 done)`)

#### Scenario: No remote lists
- **WHEN** Todex returns no lists for the authenticated user
- **THEN** the CLI SHALL print a message indicating no lists were found

### Requirement: Remote list creation
The system SHALL create todo lists through Todex using the authenticated list API.

#### Scenario: Create list
- **WHEN** the user runs `godos lists create work`
- **THEN** the CLI SHALL create a remote list named `work` and print a confirmation

#### Scenario: Duplicate list
- **WHEN** the user creates a list name that already exists remotely
- **THEN** the CLI SHALL exit non-zero with a clear duplicate-list error

### Requirement: Remote list rename
The system SHALL rename todo lists by resolving the current list name to a remote list UUID and updating that list through Todex.

#### Scenario: Rename existing list
- **WHEN** the user runs `godos lists rename work projects`
- **THEN** the CLI SHALL update the matching remote list name to `projects`

#### Scenario: Rename missing list
- **WHEN** the user runs `godos lists rename missing projects` and no remote list named `missing` exists
- **THEN** the CLI SHALL exit non-zero with a list-not-found error

### Requirement: Remote list deletion
The system SHALL delete todo lists by resolving the list name to a remote list UUID and deleting that list through Todex.

#### Scenario: Delete existing list with confirmation
- **WHEN** the user runs `godos lists delete work` and confirms the prompt
- **THEN** the CLI SHALL delete the matching remote list through Todex

#### Scenario: Delete with force flag
- **WHEN** the user runs `godos lists delete work --force`
- **THEN** the CLI SHALL delete the matching remote list without prompting

#### Scenario: Delete missing list
- **WHEN** the user deletes a list name that does not exist remotely
- **THEN** the CLI SHALL exit non-zero with a list-not-found error

### Requirement: Remote task listing
The system SHALL list tasks from Todex instead of parsing local markdown files.

#### Scenario: List tasks in default list
- **WHEN** the user runs `godos list`
- **THEN** the CLI SHALL show tasks from the remote `todo` list

#### Scenario: List tasks in named list
- **WHEN** the user runs `godos list --list work`
- **THEN** the CLI SHALL show tasks from the remote `work` list

#### Scenario: Missing list while listing tasks
- **WHEN** the user lists tasks for a list name that does not exist remotely
- **THEN** the CLI SHALL exit non-zero with a list-not-found error

### Requirement: Remote task creation
The system SHALL create tasks through Todex and associate each task with a remote list.

#### Scenario: Add task to existing list
- **WHEN** the user runs `godos add --list work "Ship API backend"` and `work` exists remotely
- **THEN** the CLI SHALL create a remote task titled `Ship API backend` in the `work` list

#### Scenario: Add task auto-creates missing list
- **WHEN** the user runs `godos add --list ideas "Try Todex"` and no remote list named `ideas` exists
- **THEN** the CLI SHALL create the remote `ideas` list and then create the task in that list

#### Scenario: Add empty task
- **WHEN** the user runs `godos add "   "`
- **THEN** the CLI SHALL reject the task before sending an API request

### Requirement: Remote task completion
The system SHALL complete tasks through Todex using the task completion endpoint and a resolved task UUID.

#### Scenario: Complete task by prefix
- **WHEN** the user runs `godos done 9f3a2c1e` and that prefix resolves to exactly one task
- **THEN** the CLI SHALL call Todex to complete the matching task

#### Scenario: Complete already completed task
- **WHEN** the user completes a task that Todex already marks completed
- **THEN** the CLI SHALL report that the task is already complete or otherwise preserve idempotent success

### Requirement: Remote task reopen
The system SHALL reopen a completed task through Todex using the task reopen endpoint and a resolved task UUID.

#### Scenario: Reopen task by prefix
- **WHEN** the user runs `godos undone 9f3a2c1e` and that prefix resolves to exactly one task
- **THEN** the CLI SHALL call Todex `POST /api/tasks/{id}/reopen` to mark the matching task incomplete

#### Scenario: Reopen unknown prefix
- **WHEN** the user runs `godos undone deadbeef` and no task matches that prefix
- **THEN** the CLI SHALL exit non-zero with a task-not-found error

### Requirement: Remote task deletion
The system SHALL delete tasks through Todex using a resolved task UUID.

#### Scenario: Remove task by prefix
- **WHEN** the user runs `godos rm a18b7d44` and that prefix resolves to exactly one task
- **THEN** the CLI SHALL delete the matching remote task through Todex

#### Scenario: Remove missing task
- **WHEN** the user runs `godos rm deadbeef` and no task matches that prefix
- **THEN** the CLI SHALL exit non-zero with a task-not-found error

### Requirement: No filesystem todo storage
The system SHALL NOT use local markdown files as the source of truth for todo lists or tasks.

#### Scenario: Todo command does not read storage directory
- **WHEN** the user runs a todo or list command after this change
- **THEN** the command SHALL use Todex API data rather than local `.md` list files
