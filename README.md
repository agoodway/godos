# godos

A command-line todo list manager. Create, organize, and track tasks across multiple markdown-backed lists.

## Install

```bash
go install github.com/goodway/godos@latest
```

Or build from source:

```bash
git clone https://github.com/goodway/godos.git
cd godos
go build -o godos .
```

## Quick Start

```bash
# Add some todos
godos add "Buy groceries"
godos add "Walk the dog"

# See your list
godos list

# Mark one done
godos done 1

# Remove one
godos rm 2
```

## Usage

### Adding Todos

```bash
# Add to the default "todo" list
godos add "Write tests"

# Add to a specific list
godos add --list work "Review pull request"
godos add --list shopping "Milk"
godos add --list shopping "Eggs"
```

### Viewing Todos

```bash
# Show the default "todo" list
godos list

# Show a specific list
godos list --list shopping

# Show all lists at once
godos list --all
```

### Completing Todos

```bash
# Mark todo #1 as done
godos done 1

# Mark a todo in a specific list
godos done 2 --list shopping
```

### Removing Todos

```bash
# Remove todo #1
godos rm 1

# Remove from a specific list
godos rm 3 --list work
```

### Managing Lists

```bash
# Show all lists with completion stats
godos lists

# Create a new empty list
godos lists create shopping

# Rename a list
godos lists rename shopping groceries

# Delete a list (prompts for confirmation)
godos lists delete groceries

# Delete without confirmation
godos lists delete groceries --force
```

### Configuration

godos stores configuration in `~/.config/godos/config.yaml` (respects `XDG_CONFIG_HOME`).

```bash
# Set the default storage directory
godos configure set default_dir ~/my-todos

# Get a config value
godos configure get default_dir

# List all config values
godos configure list

# Remove a config value
godos configure delete default_dir
```

### Storage Directory

The storage directory is resolved in this order:

1. `--dir` flag (highest priority)
2. `GODOS_DIR` environment variable
3. `default_dir` config setting
4. `~/.godos` (default)

```bash
# Use a custom directory for one command
godos --dir /tmp/todos add "Temporary task"

# Set via environment variable
export GODOS_DIR=~/work-todos
godos add "Ship feature"

# Set permanently via config
godos configure set default_dir ~/my-todos
```

### Version

```bash
godos version
```

## Storage Format

Each list is stored as a markdown file with checkbox syntax:

```markdown
- [ ] Buy milk
- [x] Buy eggs
- [ ] Buy bread
```

Lists are plain text, easy to edit by hand, and work with any markdown viewer.
