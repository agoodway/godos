package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestConfig(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
}

func TestAPIBaseURLRejectsUnsafeURLs(t *testing.T) {
	setupTestConfig(t)

	cases := []string{
		"http://api.example.com",
		"ftp://api.example.com",
		"https://user:pass@api.example.com",
		"not a url",
	}
	for _, value := range cases {
		t.Run(value, func(t *testing.T) {
			if err := Set(APIBaseURLKey, value); err != nil {
				t.Fatal(err)
			}
			_, err := APIBaseURL()
			if err == nil {
				t.Fatalf("expected %q to be rejected", value)
			}
		})
	}
}

func TestAPIBaseURLAllowsHTTPSAndLoopbackHTTP(t *testing.T) {
	setupTestConfig(t)

	for _, value := range []string{"https://api.example.com", "http://127.0.0.1:4000", "http://localhost:4000"} {
		t.Run(value, func(t *testing.T) {
			if err := Set(APIBaseURLKey, value); err != nil {
				t.Fatal(err)
			}
			got, err := APIBaseURL()
			if err != nil {
				t.Fatalf("expected %q to be allowed: %v", value, err)
			}
			if got != value {
				t.Fatalf("APIBaseURL() = %q, want %q", got, value)
			}
		})
	}
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
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Load() expected ErrNotFound, got %v", err)
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

func TestLoad_InvalidYAML(t *testing.T) {
	setupTestConfig(t)
	p := FilePath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(":\n\t- ][invalid"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for invalid YAML, got nil")
	}
	if errors.Is(err, ErrNotFound) {
		t.Error("Load() should not return ErrNotFound for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parsing config file") {
		t.Errorf("Load() error should mention parsing, got: %v", err)
	}
}

func TestSet_SpecialCharacters(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		value string
	}{
		{"colon in value", "url", "http://example.com:8080"},
		{"hash in value", "comment", "this # is a value"},
		{"quotes in value", "greeting", `say "hello"`},
		{"unicode value", "emoji", "cafe\u0301"},
		{"dots in key", "app.setting.name", "enabled"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestConfig(t)
			if err := Set(tc.key, tc.value); err != nil {
				t.Fatalf("Set(%q, %q) error: %v", tc.key, tc.value, err)
			}
			got, err := Get(tc.key)
			if err != nil {
				t.Fatalf("Get(%q) error: %v", tc.key, err)
			}
			if got != tc.value {
				t.Errorf("Get(%q) = %q, want %q", tc.key, got, tc.value)
			}
		})
	}
}

func TestAPIConfigEnvOverridesStoredValues(t *testing.T) {
	setupTestConfig(t)
	if err := Set(APIBaseURLKey, "https://stored.example"); err != nil {
		t.Fatal(err)
	}
	if err := Set(APITokenKey, "stored-token"); err != nil {
		t.Fatal(err)
	}

	t.Setenv(APIBaseURLEnv, "https://env.example")
	t.Setenv(APITokenEnv, "env-token")

	baseURL, err := APIBaseURL()
	if err != nil {
		t.Fatal(err)
	}
	if baseURL != "https://env.example" {
		t.Fatalf("expected env base URL, got %q", baseURL)
	}

	token, err := APIToken()
	if err != nil {
		t.Fatal(err)
	}
	if token != "env-token" {
		t.Fatalf("expected env token, got %q", token)
	}
}

func TestAPIConfigMissingValuesReturnClearErrors(t *testing.T) {
	setupTestConfig(t)

	if _, err := APIBaseURL(); err == nil || !strings.Contains(err.Error(), APIBaseURLKey) {
		t.Fatalf("expected missing base URL error, got %v", err)
	}
	if _, err := APIToken(); err == nil || !strings.Contains(err.Error(), "login") {
		t.Fatalf("expected missing token login guidance, got %v", err)
	}
}
