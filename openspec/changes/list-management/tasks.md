## 1. Storage Layer Extensions

- [ ] 1.1 Add `ListSummary` struct (name, total count, completed count) and `ListAll()` method to store
- [ ] 1.2 Add `CreateList(name)` method — creates empty `.md` file, errors if exists
- [ ] 1.3 Add `RenameList(old, new)` method — renames `.md` file, errors if source missing or target exists
- [ ] 1.4 Add `DeleteList(name)` method — removes `.md` file, errors if missing
- [ ] 1.5 Add list name validation function (alphanumeric, hyphens, underscores only)

## 2. CLI Commands

- [ ] 2.1 Implement `lists` root command — default action shows all lists with todo counts
- [ ] 2.2 Implement `lists create <name>` subcommand with name validation
- [ ] 2.3 Implement `lists rename <old> <new>` subcommand
- [ ] 2.4 Implement `lists delete <name>` subcommand with confirmation prompt and `--force` flag

## 3. Polish

- [ ] 3.1 Add user-friendly output messages for all list management operations
- [ ] 3.2 Verify all list commands work end-to-end with manual smoke test
