package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
	return os.MkdirAll(s.Dir, 0o755)
}

// listPath returns the file path for a named list.
func (s *Store) listPath(name string) string {
	return filepath.Join(s.Dir, name+".md")
}

// ReadList reads and parses a list file. Returns empty lines if the file doesn't exist.
func (s *Store) ReadList(name string) ([]Line, error) {
	data, err := os.ReadFile(s.listPath(name))
	if os.IsNotExist(err) {
		return nil, nil
	}
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

	target := s.listPath(name)
	content := RenderMarkdown(lines)

	tmp, err := os.CreateTemp(s.Dir, ".godos-tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, target); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

// Lists returns all list names by scanning for *.md files in the storage directory.
func (s *Store) Lists() ([]string, error) {
	entries, err := os.ReadDir(s.Dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".md") {
			names = append(names, strings.TrimSuffix(e.Name(), ".md"))
		}
	}
	return names, nil
}

// Add appends a new incomplete todo to the named list.
func (s *Store) Add(listName, text string) error {
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
