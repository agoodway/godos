## REMOVED Requirements

### Requirement: Show all lists
**Reason**: Lists are read from the Todex list API instead of scanning `.md` files.
**Migration**: See `remote-todo-storage` → "Remote list discovery", which preserves the name-and-todo-count display by counting tasks per list.

### Requirement: Create list
**Reason**: Filesystem list creation (writing a `.md` file with filename validation) is replaced by remote list creation.
**Migration**: See `remote-todo-storage` → "Remote list creation".

### Requirement: Rename list
**Reason**: Renaming a `.md` file is replaced by resolving the list name to a remote UUID and updating it through Todex.
**Migration**: See `remote-todo-storage` → "Remote list rename".

### Requirement: Delete list
**Reason**: Deleting a `.md` file is replaced by resolving the list name to a remote UUID and deleting it through Todex.
**Migration**: See `remote-todo-storage` → "Remote list deletion".
