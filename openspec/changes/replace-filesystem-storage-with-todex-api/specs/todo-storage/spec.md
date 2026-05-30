## REMOVED Requirements

### Requirement: Markdown file format
**Reason**: Todos are no longer stored as local markdown files; the Todex API is the source of truth.
**Migration**: Task state (title, completion) is now modeled by the Todex `Task` schema. See `remote-todo-storage` and `todex-api-client`. Notes retain their own markdown storage under the `notes` capability and are unaffected.

### Requirement: Storage directory management
**Reason**: Todo lists no longer write `.md` files into a storage directory, so todo-specific directory auto-creation no longer applies.
**Migration**: The storage directory is still resolved (`--dir` > `GODOS_DIR` > `default_dir` > `~/.godos`) and managed for note storage under the `notes` capability; todo/list data lives in Todex.

### Requirement: List discovery
**Reason**: Lists are discovered from the Todex list API, not by scanning `*.md` files.
**Migration**: See `remote-todo-storage` → "Remote list discovery".

### Requirement: Atomic file writes
**Reason**: Todo list mutations are performed through Todex API calls rather than local file writes, so todo-list atomic file writing no longer applies.
**Migration**: Note writes retain atomic file semantics under the `notes` capability; todo/list writes are remote API operations.
