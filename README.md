# gpeace

[![Release](https://img.shields.io/github/v/release/katolikov/gpeace?style=flat-square&color=blue)](https://github.com/katolikov/gpeace/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/katolikov/gpeace?style=flat-square&color=00ADD8)](https://go.dev/)
[![Build](https://img.shields.io/github/actions/workflow/status/katolikov/gpeace/release.yml?style=flat-square&label=build)](https://github.com/katolikov/gpeace/actions)
[![Tests](https://img.shields.io/github/actions/workflow/status/katolikov/gpeace/release.yml?style=flat-square&label=tests&color=brightgreen)](https://github.com/katolikov/gpeace/actions)
[![License](https://img.shields.io/github/license/katolikov/gpeace?style=flat-square&color=orange)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS-lightgrey?style=flat-square)](https://github.com/katolikov/gpeace/releases)

> Interactive Git merge conflict resolver for the terminal. A single compiled binary with a modern, color-coded TUI built with [Charm](https://charm.sh).

---

## Demo

When you run `gpeace` inside a Git repo with merge conflicts, you get an interactive full-screen TUI:

```
╔══════════════════════════════════════════════════════════════════╗
║  gpeace   File 1/2: src/app.py  |  Conflict 1/3               ║
╚══════════════════════════════════════════════════════════════════╝

  ╭─ Current Change (HEAD) ──────────────────────────────────────╮
  │  1 │ def greet(name):                                        │
  │  2 │     return f"Hi there, {name}! Welcome!"                │
  │  3 │                                                         │
  │  4 │ def calculate(a, b):                                    │
  │  5 │     if not isinstance(a, (int, float)):                 │
  │  6 │         raise TypeError("Must be numbers")              │
  │  7 │     return a + b                                        │
  ╰──────────────────────────────────────────────────────────────╯

  ╭─ Incoming Change (feature-branch) ──────────────────────────╮
  │  1 │ def greet(name):                                        │
  │  2 │     return f"Hey {name}, good to see you!"              │
  │  3 │                                                         │
  │  4 │ def calculate(a, b):                                    │
  │  5 │     result = a + b                                      │
  │  6 │     return result                                       │
  ╰──────────────────────────────────────────────────────────────╯

    > Accept Current Change
      Accept Incoming Change
      Accept Both Changes
      ────────────────────────
      Skip This Conflict
      Skip Entire File

  ↑/↓ navigate  enter select  q quit
```

After resolving all conflicts, you see a summary screen:

```
  gpeace - Resolution Complete

  ✓ src/app.py           3 conflicts resolved
  ✓ src/utils.go         1 conflict resolved
  ~ config.json          2 resolved, 1 skipped (not staged)
  - test/helpers.go      skipped entirely

  Press q or enter to exit.
```

## Features

- **Auto-detects** all conflicted files in the current Git repository
- **Parses** standard conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`)
- **Color-coded panels** — green for current/HEAD, blue for incoming changes
- **5 resolution options** — Accept Current, Accept Incoming, Accept Both, Skip Conflict, Skip File
- **Auto-stages** resolved files with `git add`
- **Atomic writes** — temp file + rename prevents corruption
- **Skip support** — leave markers in place for manual editing later
- **Single binary** — no runtime dependencies, just download and run

## Installation

### Download binary (recommended)

Download the latest release for your platform from the [Releases page](https://github.com/katolikov/gpeace/releases/latest):

```bash
# Linux x86_64
curl -Lo gpeace.tar.gz https://github.com/katolikov/gpeace/releases/latest/download/gpeace-linux-amd64.tar.gz
tar xzf gpeace.tar.gz && sudo mv gpeace-linux-amd64 /usr/local/bin/gpeace

# Linux ARM64
curl -Lo gpeace.tar.gz https://github.com/katolikov/gpeace/releases/latest/download/gpeace-linux-arm64.tar.gz
tar xzf gpeace.tar.gz && sudo mv gpeace-linux-arm64 /usr/local/bin/gpeace

# macOS Apple Silicon
curl -Lo gpeace.tar.gz https://github.com/katolikov/gpeace/releases/latest/download/gpeace-darwin-arm64.tar.gz
tar xzf gpeace.tar.gz && sudo mv gpeace-darwin-arm64 /usr/local/bin/gpeace

# macOS Intel
curl -Lo gpeace.tar.gz https://github.com/katolikov/gpeace/releases/latest/download/gpeace-darwin-amd64.tar.gz
tar xzf gpeace.tar.gz && sudo mv gpeace-darwin-amd64 /usr/local/bin/gpeace
```

### Build from source

```bash
go install github.com/katolikov/gpeace@latest
```

Or clone and build:

```bash
git clone https://github.com/katolikov/gpeace.git
cd gpeace
go build -o gpeace .
```

## Usage

Run inside any Git repository that has merge conflicts:

```bash
gpeace
```

That's it. The tool finds all conflicted files, walks you through each conflict one by one, and writes the resolved files when you're done.

### Controls

| Key            | Action         |
|----------------|----------------|
| `↑` / `k`     | Move cursor up |
| `↓` / `j`     | Move cursor down |
| `Enter`        | Select option  |
| `q` / `Ctrl+C` | Quit          |

### Resolution Options

| Option                   | What it does                                              |
|--------------------------|-----------------------------------------------------------|
| **Accept Current Change**  | Keep the HEAD (yours) version                            |
| **Accept Incoming Change** | Keep the incoming (theirs) version                       |
| **Accept Both Changes**    | Keep both versions concatenated                          |
| **Skip This Conflict**     | Leave conflict markers in place for manual editing later |
| **Skip Entire File**       | Skip the file entirely, move to the next one             |

### Workflow

```
  ┌─────────────┐     ┌──────────────┐     ┌──────────────┐     ┌────────────┐
  │  git merge   │────▶│   gpeace     │────▶│  resolved    │────▶│ git commit │
  │  (conflict!) │     │  (interactive)│     │  (git add)   │     │            │
  └─────────────┘     └──────────────┘     └──────────────┘     └────────────┘
```

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

## Architecture

```
gpeace/
├── main.go                 # Entry point
├── internal/
│   ├── git/git.go          # Git operations (detect conflicts, stage)
│   ├── parser/parser.go    # Conflict marker state machine
│   ├── resolver/resolver.go# File reconstruction + atomic write
│   └── tui/
│       ├── model.go        # Bubbletea TUI (Model/Update/View)
│       ├── styles.go       # Lipgloss color definitions
│       └── keys.go         # Key bindings
└── scripts/
    └── test_integration.sh # End-to-end test
```

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [bubbles](https://github.com/charmbracelet/bubbles) — Key binding helpers

## License

[Apache 2.0](LICENSE)
