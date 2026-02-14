package cmd

import (
	"bytes"
	"path/filepath"
	"sync"
	"testing"
)

// resetState resets package-level state between tests so each test
// gets a fresh store pointing at its own temp directory.
func resetState(t *testing.T) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "todos")
	dirFlag = dir
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

	_, err := executeCmd("add", "task one")
	if err != nil {
		t.Fatal(err)
	}

	_, err = executeCmd("done", "1")
	if err != nil {
		t.Fatalf("done command failed: %v", err)
	}
}

func TestDoneCmdInvalidArg(t *testing.T) {
	resetState(t)

	_, err := executeCmd("done", "abc")
	if err == nil {
		t.Error("expected error for non-numeric arg")
	}
}

func TestRmCmd(t *testing.T) {
	resetState(t)

	_, err := executeCmd("add", "task one")
	if err != nil {
		t.Fatal(err)
	}

	_, err = executeCmd("rm", "1")
	if err != nil {
		t.Fatalf("rm command failed: %v", err)
	}
}

func TestRmCmdInvalidArg(t *testing.T) {
	resetState(t)

	_, err := executeCmd("rm", "abc")
	if err == nil {
		t.Error("expected error for non-numeric arg")
	}
}

func TestListAllCmd(t *testing.T) {
	resetState(t)

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

	_, err := executeCmd("list")
	if err != nil {
		t.Fatalf("list empty failed: %v", err)
	}
}
