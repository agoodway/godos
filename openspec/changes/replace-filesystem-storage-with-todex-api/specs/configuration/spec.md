## MODIFIED Requirements

### Requirement: List all configuration values
The system SHALL accept `godos configure list` and print all key-value pairs from the config file. Secret values, including the stored Todex JWT bearer token, SHALL be masked rather than printed verbatim.

#### Scenario: Config has entries
- **WHEN** user runs `godos configure list` and the config file has entries
- **THEN** the system prints each key-value pair, one per line, in `key: value` format

#### Scenario: Config is empty or missing
- **WHEN** user runs `godos configure list` and the config file is empty or does not exist
- **THEN** the system prints a message indicating no configuration is set

#### Scenario: Stored token is masked
- **WHEN** user runs `godos configure list` and a Todex bearer token is stored
- **THEN** the system SHALL mask the token value (for example showing `api_token: ****`) instead of printing the raw token
