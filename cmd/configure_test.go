package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/goodway/godos/config"
	"github.com/spf13/pflag"
)

func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return executeCommandWithInput(t, nil, args...)
}

func executeCommandWithInput(t *testing.T, stdin io.Reader, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	if stdin != nil {
		rootCmd.SetIn(stdin)
		// Cobra doesn't propagate stdin to subcommands, so set it on all of them.
		for _, c := range rootCmd.Commands() {
			c.SetIn(stdin)
			for _, sc := range c.Commands() {
				sc.SetIn(stdin)
			}
		}
	}
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
		rootCmd.SetIn(nil)
		// Reset flags and stdin on all subcommands to prevent test contamination.
		for _, c := range rootCmd.Commands() {
			c.SetIn(nil)
			c.Flags().Visit(func(f *pflag.Flag) { f.Value.Set(f.DefValue) })
			for _, sc := range c.Commands() {
				sc.SetIn(nil)
				sc.Flags().Visit(func(f *pflag.Flag) { f.Value.Set(f.DefValue) })
			}
		}
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

func TestConfigureListMasksAPIToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if _, err := executeCommand(t, "configure", "set", "api_token", "secret-token"); err != nil {
		t.Fatal(err)
	}

	out, err := executeCommand(t, "configure", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "secret-token") {
		t.Fatalf("expected token to be masked, got %q", out)
	}
	if !strings.Contains(out, "api_token: ****") {
		t.Fatalf("expected masked token output, got %q", out)
	}
}

func TestConfigureSetMasksAPIToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	out, err := executeCommand(t, "configure", "set", config.APITokenKey, "secret-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "secret-token") {
		t.Fatalf("expected set output to redact token, got %q", out)
	}
	if !strings.Contains(out, "Set api_token = ****") {
		t.Fatalf("expected redacted set output, got %q", out)
	}
}

func TestConfigureGetAPITokenRequiresExplicitReveal(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if _, err := executeCommand(t, "configure", "set", config.APITokenKey, "secret-token"); err != nil {
		t.Fatal(err)
	}

	_, err := executeCommand(t, "configure", "get", config.APITokenKey)
	if err == nil || !strings.Contains(err.Error(), "refusing to print api_token") {
		t.Fatalf("expected refusal to print token, got %v", err)
	}

	out, err := executeCommand(t, "configure", "get", config.APITokenKey, "--show-secret")
	if err != nil {
		t.Fatalf("expected explicit reveal to succeed: %v", err)
	}
	if !strings.Contains(out, "secret-token") {
		t.Fatalf("expected revealed token, got %q", out)
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
