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

func TestValidateName(t *testing.T) {
	valid := []string{"todo", "work", "my-list", "my_list", "list123", "A", "a1-b2_c3"}
	for _, name := range valid {
		if err := ValidateName(name); err != nil {
			t.Errorf("ValidateName(%q) = %v, want nil", name, err)
		}
	}

	invalid := []string{"", "my list", "list!", "list.txt", "-start", "_start", "a/b", "a b"}
	for _, name := range invalid {
		if err := ValidateName(name); err == nil {
			t.Errorf("ValidateName(%q) = nil, want error", name)
		}
	}
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

	data, err := os.ReadFile(filepath.Join(s.Dir, "todo.md"))
	if err != nil {
		t.Fatal(err)
	}
	expected := "- [ ] first\n- [ ] second\n"
	if string(data) != expected {
		t.Errorf("file content:\ngot:  %q\nwant: %q", string(data), expected)
	}

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
	path := filepath.Join(s.Dir, "huge.md")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
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

func TestListAll_Empty(t *testing.T) {
	s := New(t.TempDir())
	summaries, err := s.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(summaries))
	}
}

func TestListAll_WithLists(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "todo.md"), []byte("- [ ] Buy milk\n- [x] Buy eggs\n- [ ] Buy bread\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "work.md"), []byte("- [ ] Review PR\n- [ ] Write docs\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	s := New(dir)
	summaries, err := s.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}

	if summaries[0].Name != "todo" {
		t.Errorf("expected first list 'todo', got %q", summaries[0].Name)
	}
	if summaries[0].Total != 3 || summaries[0].Completed != 1 {
		t.Errorf("todo: expected 3 total, 1 completed; got %d total, %d completed", summaries[0].Total, summaries[0].Completed)
	}
	if summaries[1].Name != "work" {
		t.Errorf("expected second list 'work', got %q", summaries[1].Name)
	}
	if summaries[1].Total != 2 || summaries[1].Completed != 0 {
		t.Errorf("work: expected 2 total, 0 completed; got %d total, %d completed", summaries[1].Total, summaries[1].Completed)
	}
}

func TestListAll_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "empty.md"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	s := New(dir)
	summaries, err := s.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].Total != 0 || summaries[0].Completed != 0 {
		t.Errorf("expected 0/0, got %d/%d", summaries[0].Completed, summaries[0].Total)
	}
}

func TestCreateList(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	if err := s.CreateList("shopping"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "shopping.md")); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}

	if err := s.CreateList("shopping"); err != ErrListExists {
		t.Errorf("expected ErrListExists, got %v", err)
	}
}

func TestCreateList_InvalidName(t *testing.T) {
	s := New(t.TempDir())
	if err := s.CreateList("my list!"); err != ErrInvalidName {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestRenameList(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "shopping.md"), []byte("- [ ] Milk\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	s := New(dir)
	if err := s.RenameList("shopping", "groceries"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "shopping.md")); !os.IsNotExist(err) {
		t.Error("expected old file to be removed")
	}
	data, err := os.ReadFile(filepath.Join(dir, "groceries.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "- [ ] Milk\n" {
		t.Errorf("content mismatch: %q", data)
	}
}

func TestRenameList_SourceMissing(t *testing.T) {
	s := New(t.TempDir())
	if err := s.RenameList("foo", "bar"); err != ErrListNotFound {
		t.Errorf("expected ErrListNotFound, got %v", err)
	}
}

func TestRenameList_TargetExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.md"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	s := New(dir)
	if err := s.RenameList("a", "b"); err != ErrListExists {
		t.Errorf("expected ErrListExists, got %v", err)
	}
}

func TestDeleteList(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "temp.md"), []byte("- [ ] Something\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	s := New(dir)
	if err := s.DeleteList("temp"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "temp.md")); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}
}

func TestDeleteList_Missing(t *testing.T) {
	s := New(t.TempDir())
	if err := s.DeleteList("nonexistent"); err != ErrListNotFound {
		t.Errorf("expected ErrListNotFound, got %v", err)
	}
}

func TestCountTodos(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.md"), []byte("- [ ] One\n- [x] Two\n- [X] Three\nNot a todo\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	s := New(dir)
	total, completed, err := s.CountTodos("test")
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || completed != 2 {
		t.Errorf("expected 3 total, 2 completed; got %d total, %d completed", total, completed)
	}
}

func TestCountTodos_Missing(t *testing.T) {
	s := New(t.TempDir())
	_, _, err := s.CountTodos("nonexistent")
	if err != ErrListNotFound {
		t.Errorf("expected ErrListNotFound, got %v", err)
	}
}
