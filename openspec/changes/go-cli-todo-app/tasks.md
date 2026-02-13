## 1. Project Setup

- [x] 1.1 Initialize Go module (`go mod init`) and add cobra dependency
- [x] 1.2 Create project directory structure: `cmd/`, `internal/store/`, `main.go`

## 2. Storage Layer

- [x] 2.1 Implement markdown parser — parse `- [x]` / `- [x]` lines into todo structs, preserve non-todo lines
- [x] 2.2 Implement markdown renderer — convert todo structs back to markdown lines
- [x] 2.3 Implement file I/O — read list from file, atomic write via temp-file-and-rename
- [x] 2.4 Implement storage directory management — auto-create on write, graceful empty on missing dir
- [x] 2.5 Implement list discovery — scan `*.md` files, return list names

## 3. Core Operations

- [x] 3.1 Implement add operation — append new incomplete todo, create file if missing
- [x] 3.2 Implement complete operation — mark todo done by line number, no-op if already done
- [x] 3.3 Implement remove operation — delete todo by line number, shift remaining
- [x] 3.4 Implement list operation — return numbered todos with status, empty result for missing list

## 4. CLI Commands

- [x] 4.1 Implement root command with `--dir` global flag and help output
- [x] 4.2 Implement `add` command with `--list` flag
- [x] 4.3 Implement `list` command with `--list` and `--all` flags
- [x] 4.4 Implement `done` command with `--list` flag and error handling for invalid numbers
- [x] 4.5 Implement `rm` command with `--list` flag and error handling for invalid numbers

## 5. Polish

- [x] 5.1 Add user-friendly confirmation messages and error output
- [x] 5.2 Verify all commands work end-to-end with manual smoke test
