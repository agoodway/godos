## ADDED Requirements

### Requirement: Login command
The CLI SHALL provide `godos login <email> <password>` to authenticate with Todex and persist the returned JWT bearer token.

#### Scenario: Successful login
- **WHEN** the user runs `godos login user@example.com secret` and Todex returns an auth token
- **THEN** the CLI SHALL store the token for future API requests and print a success message

#### Scenario: Failed login
- **WHEN** the user runs `godos login user@example.com wrong-password` and Todex rejects the credentials
- **THEN** the CLI SHALL not overwrite any valid stored token and SHALL exit non-zero with a clear error

### Requirement: Register command
The CLI SHALL provide `godos register <email> <password>` to create a Todex account and persist the returned JWT bearer token.

#### Scenario: Successful registration
- **WHEN** the user runs `godos register user@example.com secret` and Todex creates the account
- **THEN** the CLI SHALL store the returned token for future API requests and print a success message

#### Scenario: Registration validation failure
- **WHEN** Todex rejects the registration request
- **THEN** the CLI SHALL display the validation error and SHALL not store a new token

### Requirement: Token persistence
The system SHALL persist the JWT bearer token in local configuration with restrictive file permissions where supported by the operating system.

#### Scenario: Token saved after auth
- **WHEN** login or registration succeeds
- **THEN** subsequent API commands SHALL be able to authenticate without requiring the password again

#### Scenario: Token hidden from routine output
- **WHEN** configuration values are displayed
- **THEN** the stored bearer token SHALL be masked or omitted from routine output

### Requirement: Logout command
The CLI SHALL provide `godos logout` to end the current session by clearing the locally stored bearer token, calling `POST /api/auth/logout` when a token exists.

#### Scenario: Logout with stored token
- **WHEN** the user runs `godos logout` and a token is stored
- **THEN** the CLI SHALL call `POST /api/auth/logout` and clear the stored token

#### Scenario: Logout when API call fails
- **WHEN** the user runs `godos logout` and the `POST /api/auth/logout` request fails
- **THEN** the CLI SHALL still clear the locally stored token so the user is not left authenticated locally

#### Scenario: Logout with no stored token
- **WHEN** the user runs `godos logout` and no token is stored
- **THEN** the CLI SHALL report that no active session exists and exit without error

### Requirement: Current user validation
The system SHALL be able to validate the stored token by calling the Todex current-user endpoint.

#### Scenario: Valid token check
- **WHEN** a stored token is valid and the CLI checks the current user
- **THEN** the CLI SHALL report the authenticated user information returned by Todex

#### Scenario: Expired token check
- **WHEN** a stored token is expired or invalid
- **THEN** the CLI SHALL report that the user must login again
