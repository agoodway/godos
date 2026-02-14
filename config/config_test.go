package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestConfig(t *testing.T) (cleanup func()) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	return func() {}
}

func TestFilePath_XDGSet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	got := FilePath()
	want := filepath.Join("/custom/config", "godos", "config.yaml")
	if got != want {
		t.Errorf("FilePath() = %q, want %q", got, want)
	}
}

func TestFilePath_XDGUnset(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	got := FilePath()
	want := filepath.Join(home, ".config", "godos", "config.yaml")
	if got != want {
		t.Errorf("FilePath() = %q, want %q", got, want)
	}
}

func TestLoad_NoFile(t *testing.T) {
	setupTestConfig(t)
	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for missing file, got nil")
	}
}

func TestSaveAndLoad(t *testing.T) {
	setupTestConfig(t)
	data := map[string]string{"default_dir": "/tmp/godos", "editor": "vim"}
	if err := Save(data); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	for k, want := range data {
		if got := loaded[k]; got != want {
			t.Errorf("Load()[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestGet_Exists(t *testing.T) {
	setupTestConfig(t)
	if err := Save(map[string]string{"key1": "value1"}); err != nil {
		t.Fatal(err)
	}

	got, err := Get("key1")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got != "value1" {
		t.Errorf("Get(\"key1\") = %q, want %q", got, "value1")
	}
}

func TestGet_Missing(t *testing.T) {
	setupTestConfig(t)
	if err := Save(map[string]string{"key1": "value1"}); err != nil {
		t.Fatal(err)
	}

	_, err := Get("nonexistent")
	if err == nil {
		t.Error("Get() expected error for missing key, got nil")
	}
}

func TestGet_NoConfigFile(t *testing.T) {
	setupTestConfig(t)
	_, err := Get("anything")
	if err == nil {
		t.Error("Get() expected error when no config file exists, got nil")
	}
}

func TestSet_NewFile(t *testing.T) {
	setupTestConfig(t)
	if err := Set("default_dir", "/tmp/godos"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	got, err := Get("default_dir")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got != "/tmp/godos" {
		t.Errorf("Get(\"default_dir\") = %q, want %q", got, "/tmp/godos")
	}
}

func TestSet_OverwritePreservesOtherKeys(t *testing.T) {
	setupTestConfig(t)
	if err := Set("key1", "val1"); err != nil {
		t.Fatal(err)
	}
	if err := Set("key2", "val2"); err != nil {
		t.Fatal(err)
	}
	if err := Set("key1", "updated"); err != nil {
		t.Fatal(err)
	}

	got1, err := Get("key1")
	if err != nil {
		t.Fatal(err)
	}
	if got1 != "updated" {
		t.Errorf("Get(\"key1\") = %q, want %q", got1, "updated")
	}

	got2, err := Get("key2")
	if err != nil {
		t.Fatal(err)
	}
	if got2 != "val2" {
		t.Errorf("Get(\"key2\") = %q, want %q", got2, "val2")
	}
}
