package store

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	return New(filepath.Join(t.TempDir(), "todos"))
}

func TestAdd(t *testing.T) {
	s := newTestStore(t)
	if err := s.Add("todo", "buy milk"); err != nil {
		t.Fatal(err)
	}
	todos, err := s.ListTodos("todo")
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}
	if todos[0].Text != "buy milk" || todos[0].Done {
		t.Errorf("got %+v, want {buy milk, false}", todos[0])
	}
}

func TestAddMultiple(t *testing.T) {
	s := newTestStore(t)
	for _, text := range []string{"a", "b", "c"} {
		if err := s.Add("todo", text); err != nil {
			t.Fatal(err)
		}
	}
	todos, err := s.ListTodos("todo")
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 3 {
		t.Fatalf("expected 3 todos, got %d", len(todos))
	}
}

func TestComplete(t *testing.T) {
	s := newTestStore(t)
	if err := s.Add("todo", "first"); err != nil {
		t.Fatal(err)
	}
	if err := s.Add("todo", "second"); err != nil {
		t.Fatal(err)
	}

	text, alreadyDone, err := s.Complete("todo", 1)
	if err != nil {
		t.Fatal(err)
	}
	if text != "first" {
		t.Errorf("expected text %q, got %q", "first", text)
	}
	if alreadyDone {
		t.Error("expected alreadyDone=false")
	}

	// Verify it's actually marked done
	todos, err := s.ListTodos("todo")
	if err != nil {
		t.Fatal(err)
	}
	if !todos[0].Done {
		t.Error("todo 1 should be done")
	}
	if todos[1].Done {
		t.Error("todo 2 should not be done")
	}
}

func TestCompleteAlreadyDone(t *testing.T) {
	s := newTestStore(t)
	if err := s.Add("todo", "task"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Complete("todo", 1); err != nil {
		t.Fatal(err)
	}

	text, alreadyDone, err := s.Complete("todo", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !alreadyDone {
		t.Error("expected alreadyDone=true")
	}
	if text != "task" {
		t.Errorf("expected text %q, got %q", "task", text)
	}
}

func TestCompleteOutOfRange(t *testing.T) {
	s := newTestStore(t)
	if err := s.Add("todo", "task"); err != nil {
		t.Fatal(err)
	}

	_, _, err := s.Complete("todo", 5)
	if err == nil {
		t.Error("expected error for out-of-range")
	}
}

func TestRemove(t *testing.T) {
	s := newTestStore(t)
	for _, text := range []string{"a", "b", "c"} {
		if err := s.Add("todo", text); err != nil {
			t.Fatal(err)
		}
	}

	text, err := s.Remove("todo", 2)
	if err != nil {
		t.Fatal(err)
	}
	if text != "b" {
		t.Errorf("expected removed text %q, got %q", "b", text)
	}

	todos, err := s.ListTodos("todo")
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos after remove, got %d", len(todos))
	}
	if todos[0].Text != "a" || todos[1].Text != "c" {
		t.Errorf("expected [a, c], got [%s, %s]", todos[0].Text, todos[1].Text)
	}
}

func TestRemoveOutOfRange(t *testing.T) {
	s := newTestStore(t)
	if err := s.Add("todo", "task"); err != nil {
		t.Fatal(err)
	}

	_, err := s.Remove("todo", 5)
	if err == nil {
		t.Error("expected error for out-of-range")
	}
}

func TestLists(t *testing.T) {
	s := newTestStore(t)
	if err := s.Add("work", "task1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Add("personal", "task2"); err != nil {
		t.Fatal(err)
	}

	names, err := s.Lists()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 lists, got %d", len(names))
	}
}

func TestListsMissingDir(t *testing.T) {
	s := New(filepath.Join(t.TempDir(), "nonexistent"))
	names, err := s.Lists()
	if err != nil {
		t.Fatal(err)
	}
	if names != nil {
		t.Errorf("expected nil, got %v", names)
	}
}

func TestListTodosEmptyList(t *testing.T) {
	s := newTestStore(t)
	todos, err := s.ListTodos("missing")
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 0 {
		t.Errorf("expected 0 todos, got %d", len(todos))
	}
}

func TestPathTraversal(t *testing.T) {
	s := newTestStore(t)
	traversalNames := []string{
		"../../etc/passwd",
		"../evil",
		"foo/bar",
		"foo\\bar",
		"..",
		".",
	}
	for _, name := range traversalNames {
		err := s.Add(name, "malicious")
		if err == nil {
			t.Errorf("expected error for path traversal name %q", name)
		}
	}
}

func TestAddEmptyText(t *testing.T) {
	s := newTestStore(t)
	for _, text := range []string{"", "   ", "\t", " \n "} {
		if err := s.Add("todo", text); err == nil {
			t.Errorf("expected error for empty/whitespace text %q", text)
		}
	}
}

func TestDirPermissions(t *testing.T) {
	s := newTestStore(t)
	if err := s.Add("todo", "test"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(s.Dir)
	if err != nil {
		t.Fatal(err)
	}
	perm := info.Mode().Perm()
	if perm != 0o700 {
		t.Errorf("expected dir permissions 0700, got %04o", perm)
	}
}

func TestAtomicWrite(t *testing.T) {
	s := newTestStore(t)
	if err := s.Add("todo", "first"); err != nil {
		t.Fatal(err)
	}
	if err := s.Add("todo", "second"); err != nil {
		t.Fatal(err)
	}

	// Verify the file exists and has correct content
	data, err := os.ReadFile(filepath.Join(s.Dir, "todo.md"))
	if err != nil {
		t.Fatal(err)
	}
	expected := "- [ ] first\n- [ ] second\n"
	if string(data) != expected {
		t.Errorf("file content:\ngot:  %q\nwant: %q", string(data), expected)
	}

	// Verify no temp files were left behind (lock files are expected)
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".md" && filepath.Ext(e.Name()) != ".lock" {
			t.Errorf("unexpected file left behind: %s", e.Name())
		}
	}
}

func TestWriteListInvalidName(t *testing.T) {
	s := newTestStore(t)
	err := s.WriteList("../evil", []Line{{Todo: &Todo{Text: "hack", Done: false}}})
	if err == nil {
		t.Error("expected error for invalid list name")
	}
}

func TestWriteListReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	s := New(filepath.Join(dir, "readonly"))
	// Create the directory then make it read-only
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(s.Dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(s.Dir, 0o700) })

	err := s.WriteList("test", []Line{{Todo: &Todo{Text: "item", Done: false}}})
	if err == nil {
		t.Error("expected error writing to read-only directory")
	}
}

func TestReadListFileSizeLimit(t *testing.T) {
	s := newTestStore(t)
	if err := s.ensureDir(); err != nil {
		t.Fatal(err)
	}
	// Create a file that exceeds maxFileSize
	path := filepath.Join(s.Dir, "huge.md")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	// Write just over the limit
	data := make([]byte, maxFileSize+1)
	for i := range data {
		data[i] = 'x'
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	_, err = s.ReadList("huge")
	if err == nil {
		t.Error("expected error for oversized file")
	}
}

func TestWriteListEnsureDirFail(t *testing.T) {
	// Point at a path where the parent is a file, so MkdirAll fails
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := New(filepath.Join(blocker, "subdir"))

	err := s.WriteList("test", []Line{{Todo: &Todo{Text: "item", Done: false}}})
	if err == nil {
		t.Error("expected error when ensureDir fails")
	}
}
