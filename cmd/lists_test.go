package cmd

import (
	"strings"
	"sync"
	"testing"
)

func setupListsTestDir(t *testing.T) *remoteState {
	t.Helper()
	dirFlag = ""
	storeOnce = sync.Once{}
	todoStore = nil
	return setupRemoteCommandTest(t)
}

func TestListsCommand_NoLists(t *testing.T) {
	setupListsTestDir(t)

	out, err := executeCommand(t, "lists")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No lists found") {
		t.Errorf("expected 'No lists found', got %q", out)
	}
}

func TestListsCommand_WithLists(t *testing.T) {
	state := setupListsTestDir(t)
	state.lists = []remoteList{{ID: "11111111-1111-4111-8111-111111111111", Name: "todo"}, {ID: "22222222-1111-4111-8111-111111111111", Name: "work"}}
	state.tasks = []remoteTask{{ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", ListID: state.lists[0].ID, Title: "Task 1", Status: "active"}, {ID: "bbbbbbbb-aaaa-4aaa-8aaa-aaaaaaaaaaaa", ListID: state.lists[0].ID, Title: "Task 2", Status: "completed"}, {ID: "cccccccc-aaaa-4aaa-8aaa-aaaaaaaaaaaa", ListID: state.lists[1].ID, Title: "Meeting", Status: "active"}}

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
	state := setupListsTestDir(t)

	out, err := executeCommand(t, "lists", "create", "shopping")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Created list") {
		t.Errorf("expected creation confirmation, got %q", out)
	}

	if len(state.lists) != 1 || state.lists[0].Name != "shopping" {
		t.Errorf("expected remote list to exist: %#v", state.lists)
	}
}

func TestListsCreate_Duplicate(t *testing.T) {
	state := setupListsTestDir(t)
	state.lists = []remoteList{{ID: "11111111-1111-4111-8111-111111111111", Name: "work"}}

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
	state := setupListsTestDir(t)
	state.lists = []remoteList{{ID: "11111111-1111-4111-8111-111111111111", Name: "shopping"}}

	out, err := executeCommand(t, "lists", "rename", "shopping", "groceries")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Renamed") {
		t.Errorf("expected rename confirmation, got %q", out)
	}

	if state.lists[0].Name != "groceries" {
		t.Errorf("expected renamed remote list, got %#v", state.lists)
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
	state := setupListsTestDir(t)
	state.lists = []remoteList{{ID: "11111111-1111-4111-8111-111111111111", Name: "temp"}}

	out, err := executeCommand(t, "lists", "delete", "temp", "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Deleted") {
		t.Errorf("expected delete confirmation, got %q", out)
	}
	if len(state.lists) != 0 {
		t.Error("expected remote list to be removed")
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
	state := setupListsTestDir(t)
	state.lists = []remoteList{{ID: "11111111-1111-4111-8111-111111111111", Name: "temp"}}
	state.tasks = []remoteTask{{ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", ListID: state.lists[0].ID, Title: "Task", Status: "active"}}

	out, err := executeCommandWithInput(t, strings.NewReader("y\n"), "lists", "delete", "temp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Deleted") {
		t.Errorf("expected delete confirmation, got %q", out)
	}
	if len(state.lists) != 0 {
		t.Error("expected remote list to be removed")
	}
}

func TestListsDelete_ConfirmNo(t *testing.T) {
	state := setupListsTestDir(t)
	state.lists = []remoteList{{ID: "11111111-1111-4111-8111-111111111111", Name: "temp"}}
	state.tasks = []remoteTask{{ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", ListID: state.lists[0].ID, Title: "Task", Status: "active"}}

	out, err := executeCommandWithInput(t, strings.NewReader("n\n"), "lists", "delete", "temp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Cancelled") {
		t.Errorf("expected cancel message, got %q", out)
	}
	if len(state.lists) != 1 {
		t.Error("expected remote list to still exist")
	}
}

func TestListsDelete_ConfirmEmpty(t *testing.T) {
	state := setupListsTestDir(t)
	state.lists = []remoteList{{ID: "11111111-1111-4111-8111-111111111111", Name: "temp"}}
	state.tasks = []remoteTask{{ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", ListID: state.lists[0].ID, Title: "Task", Status: "active"}}

	out, err := executeCommandWithInput(t, strings.NewReader("\n"), "lists", "delete", "temp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Cancelled") {
		t.Errorf("expected cancel message, got %q", out)
	}
	if len(state.lists) != 1 {
		t.Error("expected remote list to still exist")
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
