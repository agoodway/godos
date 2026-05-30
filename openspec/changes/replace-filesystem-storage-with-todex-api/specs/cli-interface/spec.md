## REMOVED Requirements

### Requirement: Add command
**Reason**: The filesystem-based `godos add` behavior (appending `- [ ] <text>` to a `.md` file) is replaced by API-backed task creation.
**Migration**: See `remote-todo-storage` → "Remote task creation" (creates the task through Todex and auto-creates a missing remote list).

### Requirement: List command
**Reason**: The filesystem-based `godos list` behavior (reading `.md` files, 1-based line numbers, `--all` over local files) is replaced by API-backed task listing with task ID prefixes.
**Migration**: See `remote-todo-storage` → "Remote task listing" and `task-id-prefixes` → "Short task ID display".

### Requirement: Done command
**Reason**: The positional `godos done <number>` behavior is unsafe against a remote backend whose ordering can change between commands; it is replaced by prefix-based completion.
**Migration**: See `remote-todo-storage` → "Remote task completion" and `task-id-prefixes` → "Unique task prefix resolution" / "Positional task mutation rejected".

### Requirement: Remove command
**Reason**: The positional `godos rm <number>` behavior is unsafe against a remote backend; it is replaced by prefix-based deletion.
**Migration**: See `remote-todo-storage` → "Remote task deletion" and `task-id-prefixes` → "Unique task prefix resolution" / "Positional task mutation rejected".
