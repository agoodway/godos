package store

import (
	"os"
	"path/filepath"
	"strings"
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

func TestNotesDir(t *testing.T) {
	s := New(filepath.Join(t.TempDir(), "data"))
	got := s.NotesDir()
	want := filepath.Join(s.Dir, "notes")
	if got != want {
		t.Fatalf("NotesDir() = %q, want %q", got, want)
	}
}

func TestNotePath(t *testing.T) {
	s := New(t.TempDir())
	path, err := s.NotePath("meeting-notes")
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(s.Dir, "notes", "meeting-notes.md") {
		t.Fatalf("unexpected note path: %q", path)
	}
}

func TestNotePathInvalidName(t *testing.T) {
	s := New(t.TempDir())
	if _, err := s.NotePath("../escape"); err != ErrInvalidName {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}

func TestCreateNote(t *testing.T) {
	s := New(t.TempDir())
	if err := s.CreateNote("meeting-notes"); err != nil {
		t.Fatal(err)
	}
	path, err := s.NotePath("meeting-notes")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected note file to exist: %v", err)
	}
}

func TestCreateNoteDuplicate(t *testing.T) {
	s := New(t.TempDir())
	if err := s.CreateNote("meeting-notes"); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateNote("meeting-notes"); err != ErrNoteExists {
		t.Fatalf("expected ErrNoteExists, got %v", err)
	}
}

func TestCreateNoteInvalidName(t *testing.T) {
	s := New(t.TempDir())
	if err := s.CreateNote("invalid name!"); err != ErrInvalidName {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
	if _, err := os.Stat(s.NotesDir()); !os.IsNotExist(err) {
		t.Fatalf("expected notes dir to not be created on invalid name, stat err=%v", err)
	}
}

func TestReadNote(t *testing.T) {
	s := New(t.TempDir())
	if err := s.CreateNote("ideas"); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteNote("ideas", "line one\nline two\n"); err != nil {
		t.Fatal(err)
	}
	content, err := s.ReadNote("ideas")
	if err != nil {
		t.Fatal(err)
	}
	if content != "line one\nline two\n" {
		t.Fatalf("unexpected content: %q", content)
	}
}

func TestReadNoteMissing(t *testing.T) {
	s := New(t.TempDir())
	_, err := s.ReadNote("missing")
	if err != ErrNoteNotFound {
		t.Fatalf("expected ErrNoteNotFound, got %v", err)
	}
}

func TestWriteNote(t *testing.T) {
	s := New(t.TempDir())
	if err := s.WriteNote("ideas", "first\n"); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteNote("ideas", "second\n"); err != nil {
		t.Fatal(err)
	}
	path, err := s.NotePath("ideas")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "second\n" {
		t.Fatalf("unexpected content: %q", string(data))
	}
}

func TestWriteNoteInvalidName(t *testing.T) {
	s := New(t.TempDir())
	if err := s.WriteNote("../escape", "x"); err != ErrInvalidName {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
	if _, err := os.Stat(s.NotesDir()); !os.IsNotExist(err) {
		t.Fatalf("expected notes dir to not be created on invalid name, stat err=%v", err)
	}
}

func TestDeleteNote(t *testing.T) {
	s := New(t.TempDir())
	if err := s.CreateNote("meeting-notes"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteNote("meeting-notes"); err != nil {
		t.Fatal(err)
	}
	path, err := s.NotePath("meeting-notes")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected note file to be deleted, stat err=%v", err)
	}
}

func TestDeleteNoteMissing(t *testing.T) {
	s := New(t.TempDir())
	if err := s.DeleteNote("missing"); err != ErrNoteNotFound {
		t.Fatalf("expected ErrNoteNotFound, got %v", err)
	}
}

func TestListNotes(t *testing.T) {
	s := New(t.TempDir())
	if err := s.WriteNote("ideas", "one\n"); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteNote("meeting-notes", "two\n"); err != nil {
		t.Fatal(err)
	}
	names, err := s.ListNotes()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 notes, got %d (%v)", len(names), names)
	}
	if names[0] != "ideas" || names[1] != "meeting-notes" {
		t.Fatalf("unexpected names: %v", names)
	}
}

func TestListNotesMissingDir(t *testing.T) {
	s := New(t.TempDir())
	names, err := s.ListNotes()
	if err != nil {
		t.Fatal(err)
	}
	if names != nil {
		t.Fatalf("expected nil for missing notes dir, got %v", names)
	}
}

func TestReadNoteTooLarge(t *testing.T) {
	s := New(t.TempDir())
	if err := s.ensureNotesDir(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.NotesDir(), "huge.md")
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

	_, err = s.ReadNote("huge")
	if err == nil {
		t.Fatal("expected error for oversized note")
	}
}

func TestCreateNoteReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	s := New(dir)
	err := s.CreateNote("readonly")
	if err == nil {
		t.Fatal("expected error creating note in read-only directory")
	}
}

func TestWriteNoteReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	s := New(dir)
	err := s.WriteNote("readonly", "x")
	if err == nil {
		t.Fatal("expected error writing note in read-only directory")
	}
}

func TestReadNoteEmptyAfterCreate(t *testing.T) {
	s := New(t.TempDir())
	if err := s.CreateNote("empty"); err != nil {
		t.Fatal(err)
	}
	content, err := s.ReadNote("empty")
	if err != nil {
		t.Fatal(err)
	}
	if content != "" {
		t.Fatalf("expected empty content, got %q", content)
	}
}

func TestReadNoteInvalidName(t *testing.T) {
	s := New(t.TempDir())
	_, err := s.ReadNote("../escape")
	if err != ErrInvalidName {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}

func TestDeleteNoteInvalidName(t *testing.T) {
	s := New(t.TempDir())
	err := s.DeleteNote("../escape")
	if err != ErrInvalidName {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}

func TestNotePathNameTooLong(t *testing.T) {
	s := New(t.TempDir())
	name := strings.Repeat("a", 256)
	_, err := s.NotePath(name)
	if err != ErrNameTooLong {
		t.Fatalf("expected ErrNameTooLong, got %v", err)
	}
}

func TestWriteNoteTempCleanup(t *testing.T) {
	s := New(t.TempDir())
	if err := s.WriteNote("ideas", "one"); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteNote("ideas", "two"); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(s.NotesDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "ideas.md" {
			t.Fatalf("unexpected file left in notes dir: %s", e.Name())
		}
	}
}

func TestWriteNoteRejectsSymlinkTarget(t *testing.T) {
	s := New(t.TempDir())
	if err := s.ensureNotesDir(); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(s.NotesDir(), "evil.md")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	err := s.WriteNote("evil", "new")
	if err == nil {
		t.Fatal("expected symlink target write rejection")
	}
}

func TestReadNoteRejectsSymlinkTarget(t *testing.T) {
	s := New(t.TempDir())
	if err := s.ensureNotesDir(); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(s.NotesDir(), "evil.md")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	_, err := s.ReadNote("evil")
	if err == nil {
		t.Fatal("expected symlink target read rejection")
	}
}

func TestDeleteNoteRejectsSymlinkTarget(t *testing.T) {
	s := New(t.TempDir())
	if err := s.ensureNotesDir(); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(s.NotesDir(), "evil.md")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	err := s.DeleteNote("evil")
	if err == nil {
		t.Fatal("expected symlink target delete rejection")
	}
}

func TestListNotesFiltersNonNoteEntries(t *testing.T) {
	s := New(t.TempDir())
	if err := s.ensureNotesDir(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.NotesDir(), "valid.md"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.NotesDir(), ".hidden.md"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.NotesDir(), "draft.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(s.NotesDir(), "subdir"), 0o700); err != nil {
		t.Fatal(err)
	}

	names, err := s.ListNotes()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "valid" {
		t.Fatalf("expected only valid note, got %v", names)
	}
}
