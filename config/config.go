package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ErrNotFound is returned by Load when the config file does not exist.
var ErrNotFound = errors.New("config file not found")

// FilePath returns the path to the config file, respecting XDG_CONFIG_HOME.
func FilePath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "godos", "config.yaml")
}

// Load reads and parses the YAML config file into a map.
func Load() (map[string]string, error) {
	data, err := os.ReadFile(FilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	m := make(map[string]string)
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	return m, nil
}

// Save writes the map to the config file atomically, creating the directory if needed.
// It writes to a temporary file first, then renames to prevent partial writes.
func Save(data map[string]string) error {
	p := FilePath()
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	out, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".config-*.yaml")
	if err != nil {
		return fmt.Errorf("creating temp config file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(out); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp config file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp config file: %w", err)
	}

	if err := os.Chmod(tmpName, 0600); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("setting config file permissions: %w", err)
	}

	if err := os.Rename(tmpName, p); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("writing config file: %w", err)
	}
	return nil
}

// Get loads the config and returns the value for a key.
func Get(key string) (string, error) {
	m, err := Load()
	if err != nil {
		return "", err
	}

	v, ok := m[key]
	if !ok {
		return "", fmt.Errorf("key %q is not set", key)
	}
	return v, nil
}

// Set loads the config, sets the key-value pair, and saves.
func Set(key, value string) error {
	m, err := Load()
	if errors.Is(err, ErrNotFound) {
		// No config file yet — start fresh.
		m = make(map[string]string)
	} else if err != nil {
		return err
	}

	m[key] = value
	return Save(m)
}
