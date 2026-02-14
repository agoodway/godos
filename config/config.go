package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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
			return nil, fmt.Errorf("no configuration file found")
		}
		return nil, err
	}

	m := make(map[string]string)
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	return m, nil
}

// Save writes the map to the config file, creating the directory if needed.
func Save(data map[string]string) error {
	p := FilePath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	out, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(p, out, 0644); err != nil {
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
	if err != nil {
		// If no config file exists yet, start with an empty map.
		m = make(map[string]string)
	}

	m[key] = value
	return Save(m)
}
