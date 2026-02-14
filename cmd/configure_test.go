package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	}()
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestConfigureSet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	out, err := executeCommand(t, "configure", "set", "default_dir", "/tmp/godos")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Set default_dir = /tmp/godos") {
		t.Errorf("expected confirmation message, got %q", out)
	}
}

func TestConfigureGet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Set first
	if _, err := executeCommand(t, "configure", "set", "mykey", "myval"); err != nil {
		t.Fatal(err)
	}

	out, err := executeCommand(t, "configure", "get", "mykey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "myval") {
		t.Errorf("expected 'myval' in output, got %q", out)
	}
}

func TestConfigureGet_MissingKey(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Set one key first so the file exists
	if _, err := executeCommand(t, "configure", "set", "exists", "yes"); err != nil {
		t.Fatal(err)
	}

	_, err := executeCommand(t, "configure", "get", "nonexistent")
	if err == nil {
		t.Error("expected error for missing key, got nil")
	}
}

func TestConfigureGet_NoConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := executeCommand(t, "configure", "get", "anything")
	if err == nil {
		t.Error("expected error when no config file, got nil")
	}
}

func TestConfigureList(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if _, err := executeCommand(t, "configure", "set", "key1", "val1"); err != nil {
		t.Fatal(err)
	}
	if _, err := executeCommand(t, "configure", "set", "key2", "val2"); err != nil {
		t.Fatal(err)
	}

	out, err := executeCommand(t, "configure", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "key1: val1") {
		t.Errorf("expected 'key1: val1' in output, got %q", out)
	}
	if !strings.Contains(out, "key2: val2") {
		t.Errorf("expected 'key2: val2' in output, got %q", out)
	}
}

func TestConfigureList_NoConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	out, err := executeCommand(t, "configure", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No configuration is set") {
		t.Errorf("expected 'No configuration is set', got %q", out)
	}
}

func TestConfigure_BareCommand(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	out, err := executeCommand(t, "configure")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "configure") || !strings.Contains(out, "Usage") {
		t.Errorf("expected help/usage text, got %q", out)
	}
}

func TestConfigureSet_MissingArgs(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := executeCommand(t, "configure", "set", "onlykey")
	if err == nil {
		t.Error("expected error for missing args, got nil")
	}
}
