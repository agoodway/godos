## ADDED Requirements

### Requirement: Todex OpenAPI client
The system SHALL provide a Todex API client generated from or compatible with the OpenAPI 3.0 contract at `/Users/tbrewer/projects/goodway/todex/openapi.json`.

#### Scenario: Client generation source
- **WHEN** the API client is regenerated
- **THEN** it SHALL use the Todex OpenAPI contract as the source of truth for list, task, auth, and error schemas

#### Scenario: Generated client is buildable
- **WHEN** project tests or builds run
- **THEN** the generated or compatible Todex API client SHALL compile without requiring network access

### Requirement: Configurable API base URL
The system SHALL allow the Todex API base URL to be configured locally and overridden by environment for tests and automation.

#### Scenario: Base URL from config
- **WHEN** a user has configured an API base URL
- **THEN** API commands SHALL send requests to that base URL

#### Scenario: Base URL from environment
- **WHEN** an API base URL environment override is set
- **THEN** API commands SHALL use the environment value instead of the stored config value

#### Scenario: Missing base URL
- **WHEN** an API command needs a base URL and no base URL is available
- **THEN** the CLI SHALL return a clear configuration error

### Requirement: Bearer token requests
The system SHALL send authenticated Todex API requests with the configured JWT token in an `Authorization: Bearer <token>` header.

#### Scenario: Authenticated request includes token
- **WHEN** a stored token exists and the user runs an authenticated API command
- **THEN** the HTTP request SHALL include the bearer authorization header

#### Scenario: Missing token
- **WHEN** a user runs an authenticated API command without a stored token
- **THEN** the CLI SHALL return an error instructing the user to run `godos login` or `godos register`

### Requirement: API error normalization
The system SHALL convert Todex API error responses into concise CLI errors.

#### Scenario: Validation error response
- **WHEN** Todex returns an error response with an error message
- **THEN** the CLI SHALL display the API error message and exit non-zero

#### Scenario: Unauthorized response
- **WHEN** Todex returns `401 Unauthorized`
- **THEN** the CLI SHALL tell the user that authentication is required or expired

#### Scenario: Validation/conflict error response
- **WHEN** Todex returns a `4xx` error (such as `422` validation or `409` conflict) with an `error.message` field
- **THEN** the CLI SHALL display the API error message and exit non-zero

#### Scenario: Connection failure
- **WHEN** the CLI cannot connect to the configured Todex API
- **THEN** the CLI SHALL display a clear connection error and exit non-zero

#### Scenario: Timeout or malformed response
- **WHEN** a Todex request times out or returns a response body that cannot be parsed
- **THEN** the CLI SHALL display a clear error and exit non-zero rather than panicking
