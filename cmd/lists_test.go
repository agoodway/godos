package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupListsTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// Override the store directory by patching the env var we'll use
	t.Setenv("GODOS_DIR", dir)
	return dir
}

func TestListsCommand_NoLists(t *testing.T) {
	dir := setupListsTestDir(t)
	os.MkdirAll(dir, 0755)

	out, err := executeCommand(t, "lists")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No lists found") {
		t.Errorf("expected 'No lists found', got %q", out)
	}
}

func TestListsCommand_WithLists(t *testing.T) {
	dir := setupListsTestDir(t)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "todo.md"), []byte("- [ ] Task 1\n- [x] Task 2\n"), 0644)
	os.WriteFile(filepath.Join(dir, "work.md"), []byte("- [ ] Meeting\n"), 0644)

	out, err := executeCommand(t, "lists")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "todo") || !strings.Contains(out, "1/2 done") {
		t.Errorf("expected todo list summary, got %q", out)
	}
	if !strings.Contains(out, "work") || !strings.Contains(out, "0/1 done") {
		t.Errorf("expected work list summary, got %q", out)
	}
}

func TestListsCreate(t *testing.T) {
	dir := setupListsTestDir(t)

	out, err := executeCommand(t, "lists", "create", "shopping")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Created list") {
		t.Errorf("expected creation confirmation, got %q", out)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(dir, "shopping.md")); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}

func TestListsCreate_Duplicate(t *testing.T) {
	dir := setupListsTestDir(t)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "work.md"), []byte(""), 0644)

	_, err := executeCommand(t, "lists", "create", "work")
	if err == nil {
		t.Error("expected error for duplicate list")
	}
}

func TestListsCreate_InvalidName(t *testing.T) {
	setupListsTestDir(t)

	_, err := executeCommand(t, "lists", "create", "my list!")
	if err == nil {
		t.Error("expected error for invalid name")
	}
}

func TestListsRename(t *testing.T) {
	dir := setupListsTestDir(t)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "shopping.md"), []byte("- [ ] Milk\n"), 0644)

	out, err := executeCommand(t, "lists", "rename", "shopping", "groceries")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Renamed") {
		t.Errorf("expected rename confirmation, got %q", out)
	}

	// Old gone, new exists
	if _, err := os.Stat(filepath.Join(dir, "shopping.md")); !os.IsNotExist(err) {
		t.Error("expected old file removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "groceries.md")); err != nil {
		t.Errorf("expected new file to exist: %v", err)
	}
}

func TestListsRename_SourceMissing(t *testing.T) {
	setupListsTestDir(t)

	_, err := executeCommand(t, "lists", "rename", "foo", "bar")
	if err == nil {
		t.Error("expected error for missing source")
	}
}

func TestListsDelete_Force(t *testing.T) {
	dir := setupListsTestDir(t)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "temp.md"), []byte("- [ ] Something\n"), 0644)

	out, err := executeCommand(t, "lists", "delete", "temp", "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Deleted") {
		t.Errorf("expected delete confirmation, got %q", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "temp.md")); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}
}

func TestListsDelete_Missing(t *testing.T) {
	setupListsTestDir(t)

	_, err := executeCommand(t, "lists", "delete", "nonexistent", "--force")
	if err == nil {
		t.Error("expected error for missing list")
	}
}

func TestListsDelete_ConfirmYes(t *testing.T) {
	dir := setupListsTestDir(t)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "temp.md"), []byte("- [ ] Task\n"), 0644)

	out, err := executeCommandWithInput(t, strings.NewReader("y\n"), "lists", "delete", "temp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Deleted") {
		t.Errorf("expected delete confirmation, got %q", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "temp.md")); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}
}

func TestListsDelete_ConfirmNo(t *testing.T) {
	dir := setupListsTestDir(t)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "temp.md"), []byte("- [ ] Task\n"), 0644)

	out, err := executeCommandWithInput(t, strings.NewReader("n\n"), "lists", "delete", "temp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Cancelled") {
		t.Errorf("expected cancel message, got %q", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "temp.md")); err != nil {
		t.Error("expected file to still exist")
	}
}

func TestListsDelete_ConfirmEmpty(t *testing.T) {
	dir := setupListsTestDir(t)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "temp.md"), []byte("- [ ] Task\n"), 0644)

	out, err := executeCommandWithInput(t, strings.NewReader("\n"), "lists", "delete", "temp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Cancelled") {
		t.Errorf("expected cancel message, got %q", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "temp.md")); err != nil {
		t.Error("expected file to still exist")
	}
}

func TestListsHelp(t *testing.T) {
	out, err := executeCommand(t, "lists", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "create") || !strings.Contains(out, "rename") || !strings.Contains(out, "delete") {
		t.Errorf("expected subcommands in help, got %q", out)
	}
}
