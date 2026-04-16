# gpeace

Interactive Git merge conflict resolver for the terminal. A single compiled binary with a modern, color-coded TUI.

## Features

- Detects all conflicted files in the current Git repository
- Parses standard conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`)
- Displays each conflict with color-coded panels (green for current/HEAD, blue for incoming)
- Interactive resolution: Accept Current, Accept Incoming, Accept Both, Skip Conflict, Skip File
- Automatically stages resolved files with `git add`
- Atomic file writes to prevent corruption
- Skipped conflicts preserve original markers for manual editing later

## Build

```bash
go build -o gpeace .
```

## Install

```bash
go install github.com/katolikov/gpeace@latest
```

Or copy the binary to your PATH:

```bash
go build -o gpeace .
sudo cp gpeace /usr/local/bin/
```

## Usage

Run inside any Git repository that has merge conflicts:

```bash
gpeace
```

### Controls

| Key          | Action              |
|--------------|---------------------|
| `Up` / `k`   | Move cursor up      |
| `Down` / `j` | Move cursor down    |
| `Enter`      | Select option       |
| `q` / `Ctrl+C` | Quit              |

### Resolution Options

| Option                 | Behavior                                          |
|------------------------|---------------------------------------------------|
| Accept Current Change  | Keep the HEAD (yours) version                     |
| Accept Incoming Change | Keep the incoming (theirs) version                |
| Accept Both Changes    | Keep both versions concatenated                   |
| Skip This Conflict     | Leave conflict markers in place for manual editing|
| Skip Entire File       | Skip the file entirely, move to the next one      |

## Testing

### Unit tests

```bash
go test ./... -v
```

### Integration test

Creates a real Git repo with merge conflicts and verifies the parser:

```bash
bash scripts/test_integration.sh
```

The script also prints instructions for manual interactive testing.

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [bubbles](https://github.com/charmbracelet/bubbles) - Key binding helpers

## License

Apache 2.0
