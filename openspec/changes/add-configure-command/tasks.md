## 1. Config Package

- [ ] 1.1 Create `config/config.go` with `FilePath()` function that resolves `$XDG_CONFIG_HOME/godos/config.yaml` (fallback `~/.config/godos/config.yaml`)
- [ ] 1.2 Implement `Load()` function that reads and parses the YAML config file into `map[string]string`
- [ ] 1.3 Implement `Save(data map[string]string)` function that writes the map to the config file, creating the directory if needed
- [ ] 1.4 Implement `Get(key string) (string, error)` that loads config and returns the value for a key
- [ ] 1.5 Implement `Set(key, value string) error` that loads config, sets the key, and saves
- [ ] 1.6 Add `gopkg.in/yaml.v3` dependency

## 2. Configure Command

- [ ] 2.1 Create `cmd/configure.go` with parent `configure` command registered on `rootCmd`
- [ ] 2.2 Implement `configure set <key> <value>` subcommand that calls `config.Set()` and prints confirmation
- [ ] 2.3 Implement `configure get <key>` subcommand that calls `config.Get()` and prints the value or an error
- [ ] 2.4 Implement `configure list` subcommand that calls `config.Load()` and prints all key-value pairs

## 3. Testing

- [ ] 3.1 Write tests for `config` package: `FilePath()`, `Load()`, `Save()`, `Get()`, `Set()` using a temp directory
- [ ] 3.2 Write tests for configure command: verify set/get/list output and error cases
