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
	s.Add("todo", "first")
	s.Add("todo", "second")

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
	todos, _ := s.ListTodos("todo")
	if !todos[0].Done {
		t.Error("todo 1 should be done")
	}
	if todos[1].Done {
		t.Error("todo 2 should not be done")
	}
}

func TestCompleteAlreadyDone(t *testing.T) {
	s := newTestStore(t)
	s.Add("todo", "task")
	s.Complete("todo", 1)

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
	s.Add("todo", "task")

	_, _, err := s.Complete("todo", 5)
	if err == nil {
		t.Error("expected error for out-of-range")
	}
}

func TestRemove(t *testing.T) {
	s := newTestStore(t)
	s.Add("todo", "a")
	s.Add("todo", "b")
	s.Add("todo", "c")

	text, err := s.Remove("todo", 2)
	if err != nil {
		t.Fatal(err)
	}
	if text != "b" {
		t.Errorf("expected removed text %q, got %q", "b", text)
	}

	todos, _ := s.ListTodos("todo")
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos after remove, got %d", len(todos))
	}
	if todos[0].Text != "a" || todos[1].Text != "c" {
		t.Errorf("expected [a, c], got [%s, %s]", todos[0].Text, todos[1].Text)
	}
}

func TestRemoveOutOfRange(t *testing.T) {
	s := newTestStore(t)
	s.Add("todo", "task")

	_, err := s.Remove("todo", 5)
	if err == nil {
		t.Error("expected error for out-of-range")
	}
}

func TestLists(t *testing.T) {
	s := newTestStore(t)
	s.Add("work", "task1")
	s.Add("personal", "task2")

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
	s.Add("todo", "first")
	s.Add("todo", "second")

	// Verify the file exists and has correct content
	data, err := os.ReadFile(filepath.Join(s.Dir, "todo.md"))
	if err != nil {
		t.Fatal(err)
	}
	expected := "- [ ] first\n- [ ] second\n"
	if string(data) != expected {
		t.Errorf("file content:\ngot:  %q\nwant: %q", string(data), expected)
	}

	// Verify no temp files were left behind
	entries, _ := os.ReadDir(s.Dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".md" {
			t.Errorf("unexpected file left behind: %s", e.Name())
		}
	}
}
