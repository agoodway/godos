## REMOVED Requirements

### Requirement: Add todo
**Reason**: Appending an incomplete todo to a local list file is replaced by API task creation.
**Migration**: See `remote-todo-storage` → "Remote task creation".

### Requirement: Complete todo
**Reason**: Completing a todo by 1-based line number is replaced by API completion against a prefix-resolved task UUID.
**Migration**: See `remote-todo-storage` → "Remote task completion" and `task-id-prefixes`.

### Requirement: Remove todo
**Reason**: Removing a todo by 1-based line number is replaced by API deletion against a prefix-resolved task UUID.
**Migration**: See `remote-todo-storage` → "Remote task deletion" and `task-id-prefixes`.

### Requirement: List todos
**Reason**: Reading todos with 1-based numbers from a local file is replaced by API task listing with task ID prefixes.
**Migration**: See `remote-todo-storage` → "Remote task listing" and `task-id-prefixes` → "Short task ID display".
