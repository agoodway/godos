## REMOVED Requirements

### Requirement: Note storage directory
**Reason**: Notes are now stored in Todex and commands no longer use local markdown files as the source of truth.
**Migration**: Use remote note commands backed by the configured Todex API. Existing local files are not migrated by this change.

### Requirement: Create note
**Reason**: Note creation now creates a Todex note instead of an empty local `.md` file.
**Migration**: Use `godos note add <title>` to create a remote note.

### Requirement: Read note
**Reason**: Note retrieval now fetches a Todex note by ID prefix instead of reading a local file by name.
**Migration**: Use `godos note show <note-id-prefix>` after listing notes.

### Requirement: Write note
**Reason**: Note body updates now PATCH Todex notes instead of atomically writing local files.
**Migration**: Use `godos note edit <note-id-prefix>` to edit the remote note body through the configured editor.

### Requirement: Delete note
**Reason**: Note removal now soft-deletes a Todex note instead of removing a local file.
**Migration**: Use `godos note rm <note-id-prefix>` to soft-delete a remote note, and `godos note restore <note-id-prefix>` to restore it.

### Requirement: List notes
**Reason**: Note listing now queries Todex instead of scanning the local notes directory.
**Migration**: Use `godos notes` to list remote notes and their ID prefixes.
