## ADDED Requirements

### Requirement: Set a configuration value
The system SHALL accept `godos configure set <key> <value>` and persist the key-value pair to the config file. If the config file or its parent directory does not exist, the system SHALL create them.

#### Scenario: Set a new key
- **WHEN** user runs `godos configure set default_dir /tmp/godos`
- **THEN** the config file at `~/.config/godos/config.yaml` contains `default_dir: /tmp/godos`

#### Scenario: Overwrite an existing key
- **WHEN** user runs `godos configure set default_dir /new/path` and `default_dir` already exists in config
- **THEN** the value is updated to `/new/path` and other keys are preserved

#### Scenario: Config directory does not exist
- **WHEN** user runs `godos configure set default_dir /tmp/godos` and `~/.config/godos/` does not exist
- **THEN** the directory is created and the config file is written

#### Scenario: Missing arguments
- **WHEN** user runs `godos configure set` with fewer than 2 arguments
- **THEN** the system prints a usage error and exits with a non-zero code

### Requirement: Get a configuration value
The system SHALL accept `godos configure get <key>` and print the value for that key to stdout.

#### Scenario: Key exists
- **WHEN** user runs `godos configure get default_dir` and the key is set
- **THEN** the system prints the value to stdout

#### Scenario: Key does not exist
- **WHEN** user runs `godos configure get nonexistent_key`
- **THEN** the system prints an error message indicating the key is not set and exits with a non-zero code

#### Scenario: Config file does not exist
- **WHEN** user runs `godos configure get default_dir` and no config file exists
- **THEN** the system prints an error message indicating no configuration exists and exits with a non-zero code

### Requirement: List all configuration values
The system SHALL accept `godos configure list` and print all key-value pairs from the config file.

#### Scenario: Config has entries
- **WHEN** user runs `godos configure list` and the config file has entries
- **THEN** the system prints each key-value pair, one per line, in `key: value` format

#### Scenario: Config is empty or missing
- **WHEN** user runs `godos configure list` and the config file is empty or does not exist
- **THEN** the system prints a message indicating no configuration is set

### Requirement: Config file location respects XDG
The system SHALL use `$XDG_CONFIG_HOME/godos/config.yaml` as the config file path. If `XDG_CONFIG_HOME` is not set, the system SHALL fall back to `~/.config/godos/config.yaml`.

#### Scenario: XDG_CONFIG_HOME is set
- **WHEN** `XDG_CONFIG_HOME` is set to `/custom/config`
- **THEN** the config file is read from and written to `/custom/config/godos/config.yaml`

#### Scenario: XDG_CONFIG_HOME is not set
- **WHEN** `XDG_CONFIG_HOME` is not set
- **THEN** the config file is read from and written to `~/.config/godos/config.yaml`

### Requirement: Config package provides programmatic access
The system SHALL provide a `config` package that other commands can import to read configuration values via `config.Get(key)` and resolve the config file path via `config.FilePath()`.

#### Scenario: Another command reads default_dir
- **WHEN** a command calls `config.Get("default_dir")` and the key is set in the config file
- **THEN** the function returns the value and no error

#### Scenario: Another command reads a missing key
- **WHEN** a command calls `config.Get("missing_key")` and the key is not in the config file
- **THEN** the function returns an empty string and an error
