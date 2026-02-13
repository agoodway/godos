## ADDED Requirements

### Requirement: Lists command group
The CLI SHALL provide a `godos lists` command group. Running `godos lists` with no subcommand SHALL display all lists with todo counts. Subcommands `create`, `rename`, and `delete` SHALL be available.

#### Scenario: Lists help output
- **WHEN** user runs `godos lists --help`
- **THEN** the CLI SHALL display the available subcommands: create, rename, delete

#### Scenario: Lists command default action
- **WHEN** user runs `godos lists`
- **THEN** the CLI SHALL display all lists (equivalent to listing without a subcommand)
