package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

var (
	// ErrListExists is returned when creating a list that already exists.
	ErrListExists = errors.New("list already exists")
	// ErrListNotFound is returned when operating on a list that does not exist.
	ErrListNotFound = errors.New("list not found")
	// ErrNoteExists is returned when creating a note that already exists.
	ErrNoteExists = errors.New("note already exists")
	// ErrNoteNotFound is returned when operating on a note that does not exist.
	ErrNoteNotFound = errors.New("note not found")
	// ErrInvalidName is returned when a name contains invalid characters.
	ErrInvalidName = errors.New("invalid name: use only letters, numbers, hyphens, and underscores")
	// ErrNameTooLong is returned when a name exceeds the maximum length.
	ErrNameTooLong = errors.New("name too long: maximum 255 characters")
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// ValidateName checks if a name is valid for use as a filename.
// Names must contain only alphanumeric characters, hyphens, and underscores,
// and must start with an alphanumeric character.
func ValidateName(name string) error {
	if len(name) > 255 {
		return ErrNameTooLong
	}
	if !validName.MatchString(name) {
		return ErrInvalidName
	}
	return nil
}

// ListSummary holds per-list metadata for display.
type ListSummary struct {
	Name      string
	Total     int
	Completed int
}

// Store manages todo lists stored as markdown files in a directory.
type Store struct {
	Dir string
}

// New creates a Store for the given directory.
func New(dir string) *Store {
	return &Store{Dir: dir}
}

// ensureDir creates the storage directory if it does not exist.
func (s *Store) ensureDir() error {
	return os.MkdirAll(s.Dir, 0o700)
}

// lockList acquires an exclusive file lock for the named list.
// Returns an unlock function that must be called when done.
func (s *Store) lockList(name string) (unlock func(), err error) {
	if err := ValidateName(name); err != nil {
		return nil, err
	}
	if err := s.ensureDir(); err != nil {
		return nil, fmt.Errorf("creating storage directory: %w", err)
	}
	lockPath := filepath.Join(s.Dir, "."+name+".lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("opening lock file: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("acquiring lock: %w", err)
	}
	return func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}

// listPath returns the file path for a named list.
func (s *Store) listPath(name string) (string, error) {
	if err := ValidateName(name); err != nil {
		return "", err
	}

	p := filepath.Join(s.Dir, name+".md")
	absDir, err := filepath.Abs(s.Dir)
	if err != nil {
		return "", fmt.Errorf("resolving storage directory: %w", err)
	}
	absP, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolving list path: %w", err)
	}
	if !strings.HasPrefix(absP, absDir+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid list name %q: resolved path escapes storage directory", name)
	}

	if realPath, err := filepath.EvalSymlinks(p); err == nil {
		realDir, err2 := filepath.EvalSymlinks(s.Dir)
		if err2 != nil {
			return "", fmt.Errorf("resolving storage directory symlinks: %w", err2)
		}
		if !strings.HasPrefix(realPath, realDir+string(filepath.Separator)) {
			return "", fmt.Errorf("invalid list name %q: resolved path escapes storage directory via symlink", name)
		}
	}

	return p, nil
}

// NotesDir returns the notes storage directory.
func (s *Store) NotesDir() string {
	return filepath.Join(s.Dir, "notes")
}

func (s *Store) ensureNotesDir() error {
	notesDir := s.NotesDir()
	if info, err := os.Lstat(notesDir); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("notes directory must not be a symlink")
		}
		if !info.IsDir() {
			return fmt.Errorf("notes path is not a directory")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking notes directory: %w", err)
	}

	if err := os.MkdirAll(notesDir, 0o700); err != nil {
		return fmt.Errorf("creating notes directory: %w", err)
	}

	if info, err := os.Lstat(notesDir); err != nil {
		return fmt.Errorf("checking notes directory after create: %w", err)
	} else if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("notes directory must not be a symlink")
	}

	return nil
}

// NotePath returns the file path for a validated note name.
func (s *Store) NotePath(name string) (string, error) {
	if err := ValidateName(name); err != nil {
		return "", err
	}

	notesDir := s.NotesDir()
	p := filepath.Join(notesDir, name+".md")
	absDir, err := filepath.Abs(notesDir)
	if err != nil {
		return "", fmt.Errorf("resolving notes directory: %w", err)
	}
	absP, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolving note path: %w", err)
	}
	if !strings.HasPrefix(absP, absDir+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid note name %q: resolved path escapes notes directory", name)
	}

	if realPath, err := filepath.EvalSymlinks(p); err == nil {
		realDir, err2 := filepath.EvalSymlinks(notesDir)
		if err2 != nil {
			return "", fmt.Errorf("resolving notes directory symlinks: %w", err2)
		}
		if !strings.HasPrefix(realPath, realDir+string(filepath.Separator)) {
			return "", fmt.Errorf("invalid note name %q: resolved path escapes notes directory via symlink", name)
		}
	}

	return p, nil
}

// CreateNote creates an empty note file at notes/<name>.md.
func (s *Store) CreateNote(name string) error {
	path, err := s.NotePath(name)
	if err != nil {
		return err
	}

	if err := s.ensureNotesDir(); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrNoteExists
		}
		return fmt.Errorf("creating note: %w", err)
	}
	return f.Close()
}

// ReadNote returns the full contents of a note.
func (s *Store) ReadNote(name string) (string, error) {
	path, err := s.NotePath(name)
	if err != nil {
		return "", err
	}

	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrNoteNotFound
	}
	if err != nil {
		return "", fmt.Errorf("reading note metadata: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("refusing to read symlinked note %q", name)
	}

	statInfo, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrNoteNotFound
	}
	if err != nil {
		return "", fmt.Errorf("stat note: %w", err)
	}
	if statInfo.Size() > maxFileSize {
		return "", fmt.Errorf("note file %q is too large (%d bytes, max %d)", name, statInfo.Size(), maxFileSize)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrNoteNotFound
		}
		return "", fmt.Errorf("reading note: %w", err)
	}
	return string(data), nil
}

// WriteNote writes note content atomically to notes/<name>.md.
func (s *Store) WriteNote(name string, content string) error {
	target, err := s.NotePath(name)
	if err != nil {
		return err
	}

	if err := s.ensureNotesDir(); err != nil {
		return err
	}

	if info, err := os.Lstat(target); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to write symlinked note %q", name)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking existing note: %w", err)
	}

	tmp, err := os.CreateTemp(s.NotesDir(), ".godos-note-tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp note file: %w", err)
	}
	tmpName := tmp.Name()

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("setting temp note file permissions: %w", err)
	}

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp note file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("syncing temp note file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp note file: %w", err)
	}
	if err := os.Rename(tmpName, target); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp note file: %w", err)
	}

	if dir, err := os.Open(s.NotesDir()); err == nil {
		dir.Sync()
		dir.Close()
	}

	return nil
}

// DeleteNote removes a note file.
func (s *Store) DeleteNote(name string) error {
	path, err := s.NotePath(name)
	if err != nil {
		return err
	}

	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return ErrNoteNotFound
	}
	if err != nil {
		return fmt.Errorf("reading note metadata: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to delete symlinked note %q", name)
	}

	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNoteNotFound
		}
		return fmt.Errorf("deleting note: %w", err)
	}
	return nil
}

// ListNotes returns all note names from notes/*.md.
func (s *Store) ListNotes() ([]string, error) {
	entries, err := os.ReadDir(s.NotesDir())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if strings.HasSuffix(e.Name(), ".md") {
			name := strings.TrimSuffix(e.Name(), ".md")
			if ValidateName(name) == nil {
				names = append(names, name)
			}
		}
	}
	return names, nil
}

// maxFileSize is the maximum list or note file size accepted for reads (10MB).
const maxFileSize = 10 * 1024 * 1024

// ReadList reads and parses a list file. Returns empty lines if the file doesn't exist.
func (s *Store) ReadList(name string) ([]Line, error) {
	path, err := s.listPath(name)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("list file %q is too large (%d bytes, max %d)", name, info.Size(), maxFileSize)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseMarkdown(string(data)), nil
}

// WriteList writes lines to a list file using atomic write (temp file + rename).
func (s *Store) WriteList(name string, lines []Line) error {
	if err := s.ensureDir(); err != nil {
		return fmt.Errorf("creating storage directory: %w", err)
	}

	target, err := s.listPath(name)
	if err != nil {
		return err
	}
	content := RenderMarkdown(lines)

	tmp, err := os.CreateTemp(s.Dir, ".godos-tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("setting temp file permissions: %w", err)
	}

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("syncing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, target); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	if dir, err := os.Open(s.Dir); err == nil {
		dir.Sync()
		dir.Close()
	}
	return nil
}

// Lists returns all list names by scanning for *.md files in the storage directory.
func (s *Store) Lists() ([]string, error) {
	entries, err := os.ReadDir(s.Dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if strings.HasSuffix(e.Name(), ".md") {
			name := strings.TrimSuffix(e.Name(), ".md")
			if ValidateName(name) == nil {
				names = append(names, name)
			}
		}
	}
	return names, nil
}

// ListAll scans *.md files in the storage directory and returns a summary for each.
func (s *Store) ListAll() ([]ListSummary, error) {
	names, err := s.Lists()
	if err != nil {
		return nil, err
	}

	summaries := make([]ListSummary, 0, len(names))
	for _, name := range names {
		total, completed, err := s.CountTodos(name)
		if err != nil {
			return nil, fmt.Errorf("reading list %q: %w", name, err)
		}
		summaries = append(summaries, ListSummary{
			Name:      name,
			Total:     total,
			Completed: completed,
		})
	}
	return summaries, nil
}

// CreateList creates an empty list file. Returns ErrListExists if the file already exists.
func (s *Store) CreateList(name string) error {
	if err := s.ensureDir(); err != nil {
		return fmt.Errorf("creating storage directory: %w", err)
	}

	path, err := s.listPath(name)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrListExists
		}
		return fmt.Errorf("creating list: %w", err)
	}
	return f.Close()
}

// RenameList renames a list file. Returns ErrListNotFound if the source doesn't exist,
// or ErrListExists if the target already exists.
func (s *Store) RenameList(oldName, newName string) error {
	oldPath, err := s.listPath(oldName)
	if err != nil {
		return err
	}
	newPath, err := s.listPath(newName)
	if err != nil {
		return err
	}

	// Link fails atomically if the target already exists.
	if err := os.Link(oldPath, newPath); err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrListExists
		}
		if errors.Is(err, os.ErrNotExist) {
			return ErrListNotFound
		}
		return fmt.Errorf("renaming list: %w", err)
	}
	return os.Remove(oldPath)
}

// DeleteList removes a list file. Returns ErrListNotFound if it doesn't exist.
func (s *Store) DeleteList(name string) error {
	path, err := s.listPath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrListNotFound
		}
		return fmt.Errorf("deleting list: %w", err)
	}
	return nil
}

// CountTodos returns the total and completed todo count for a named list.
func (s *Store) CountTodos(name string) (total, completed int, err error) {
	path, err := s.listPath(name)
	if err != nil {
		return 0, 0, err
	}
	if _, statErr := os.Stat(path); statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return 0, 0, ErrListNotFound
		}
		return 0, 0, statErr
	}
	lines, err := s.ReadList(name)
	if err != nil {
		return 0, 0, err
	}
	for _, line := range lines {
		if line.Todo != nil {
			total++
			if line.Todo.Done {
				completed++
			}
		}
	}
	return total, completed, nil
}

// Add appends a new incomplete todo to the named list.
func (s *Store) Add(listName, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("todo text must not be empty")
	}

	unlock, err := s.lockList(listName)
	if err != nil {
		return err
	}
	defer unlock()

	lines, err := s.ReadList(listName)
	if err != nil {
		return err
	}
	lines = append(lines, Line{Todo: &Todo{Text: text, Done: false}})
	return s.WriteList(listName, lines)
}

// Complete marks the nth todo (1-based) as done. Returns the todo text.
// Returns an error if n is out of range. Returns (text, alreadyDone, err).
func (s *Store) Complete(listName string, n int) (string, bool, error) {
	unlock, err := s.lockList(listName)
	if err != nil {
		return "", false, err
	}
	defer unlock()

	lines, err := s.ReadList(listName)
	if err != nil {
		return "", false, err
	}

	idx := todoIndex(lines, n)
	if idx == -1 {
		count := len(ExtractTodos(lines))
		return "", false, fmt.Errorf("todo #%d not found (list has %d todos)", n, count)
	}

	todo := lines[idx].Todo
	if todo.Done {
		return todo.Text, true, nil
	}

	todo.Done = true
	if err := s.WriteList(listName, lines); err != nil {
		return "", false, err
	}
	return todo.Text, false, nil
}

// Remove removes the nth todo (1-based) from the list. Returns the removed todo text.
func (s *Store) Remove(listName string, n int) (string, error) {
	unlock, err := s.lockList(listName)
	if err != nil {
		return "", err
	}
	defer unlock()

	lines, err := s.ReadList(listName)
	if err != nil {
		return "", err
	}

	idx := todoIndex(lines, n)
	if idx == -1 {
		count := len(ExtractTodos(lines))
		return "", fmt.Errorf("todo #%d not found (list has %d todos)", n, count)
	}

	text := lines[idx].Todo.Text
	lines = append(lines[:idx], lines[idx+1:]...)
	if err := s.WriteList(listName, lines); err != nil {
		return "", err
	}
	return text, nil
}

// ListTodos returns all todos in the named list.
func (s *Store) ListTodos(listName string) ([]Todo, error) {
	lines, err := s.ReadList(listName)
	if err != nil {
		return nil, err
	}
	return ExtractTodos(lines), nil
}
