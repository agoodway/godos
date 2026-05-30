package cmd

import (
	"bytes"
	"sync"
	"testing"
)

// resetState resets package-level state between tests so each test
// gets a fresh store pointing at its own temp directory.
func resetState(t *testing.T) {
	t.Helper()
	dirFlag = ""
	storeOnce = sync.Once{}
	todoStore = nil
}

// executeCmd runs rootCmd with the given args and returns stdout output and any error.
func executeCmd(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestVersionCmd(t *testing.T) {
	resetState(t)
	_, err := executeCmd("version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
}

func TestAddAndListCmd(t *testing.T) {
	resetState(t)
	setupRemoteCommandTest(t)

	_, err := executeCmd("add", "buy milk")
	if err != nil {
		t.Fatalf("add command failed: %v", err)
	}

	_, err = executeCmd("add", "walk dog")
	if err != nil {
		t.Fatalf("add second item failed: %v", err)
	}

	_, err = executeCmd("list")
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}

func TestDoneCmd(t *testing.T) {
	resetState(t)
	state := setupRemoteCommandTest(t)
	state.lists = []remoteList{{ID: "11111111-1111-4111-8111-111111111111", Name: "todo"}}
	state.tasks = []remoteTask{{ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", ListID: state.lists[0].ID, Title: "task one", Status: "active"}}

	_, err := executeCmd("done", "aaaaaaaa")
	if err != nil {
		t.Fatalf("done command failed: %v", err)
	}
}

func TestDoneCmdInvalidArg(t *testing.T) {
	resetState(t)
	setupRemoteCommandTest(t)

	_, err := executeCmd("done", "3")
	if err == nil {
		t.Error("expected error for numeric positional arg")
	}
}

func TestRmCmd(t *testing.T) {
	resetState(t)
	state := setupRemoteCommandTest(t)
	state.lists = []remoteList{{ID: "11111111-1111-4111-8111-111111111111", Name: "todo"}}
	state.tasks = []remoteTask{{ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", ListID: state.lists[0].ID, Title: "task one", Status: "active"}}

	_, err := executeCmd("rm", "aaaaaaaa")
	if err != nil {
		t.Fatalf("rm command failed: %v", err)
	}
}

func TestRmCmdInvalidArg(t *testing.T) {
	resetState(t)
	setupRemoteCommandTest(t)

	_, err := executeCmd("rm", "3")
	if err == nil {
		t.Error("expected error for numeric positional arg")
	}
}

func TestListAllCmd(t *testing.T) {
	resetState(t)
	setupRemoteCommandTest(t)

	_, err := executeCmd("add", "--list", "work", "task A")
	if err != nil {
		t.Fatal(err)
	}
	_, err = executeCmd("add", "--list", "personal", "task B")
	if err != nil {
		t.Fatal(err)
	}

	_, err = executeCmd("list", "--all")
	if err != nil {
		t.Fatalf("list --all failed: %v", err)
	}
}

func TestListEmptyCmd(t *testing.T) {
	resetState(t)
	state := setupRemoteCommandTest(t)
	state.lists = []remoteList{{ID: "11111111-1111-4111-8111-111111111111", Name: "todo"}}

	_, err := executeCmd("list")
	if err != nil {
		t.Fatalf("list empty failed: %v", err)
	}
}
