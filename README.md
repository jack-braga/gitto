# gitto

A multi-repo Git TUI workspace manager built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Run `gitto` in any parent directory to see the git status of all child repos at a glance — then stage, commit, push, pull, branch, stash, and more without leaving your terminal.

## Why?

Tools like `lazygit` and `tig` are great but operate on a single repo. If you work across multiple related repositories, you're constantly `cd`-ing between them. Gitto gives you one interface to manage them all.

## Features

- **Overview mode** — See every repo's branch, change count, and ahead/behind status at a glance
- **Drill-in mode** — Full source control for a single repo (staged/unstaged files, stashes, commit history)
- **Three views** — Source control, file explorer with git status annotations, scrollable commit history
- **Branch picker** — List, switch, create, delete branches with fuzzy filtering
- **Stash management** — Save, pop, apply, drop stashes
- **Background polling** — Status refreshes automatically every 5 seconds (configurable)
- **Editor integration** — Open files in your preferred editor (`$EDITOR`, VS Code, vim, etc.)
- **Responsive** — Adapts to terminal width, scrollable viewport, min-size guard

## Install

```bash
# From source
go install github.com/jack-braga/gitto@latest

# Or clone and build
git clone https://github.com/jack-braga/gitto.git
cd gitto
make build
```

Requires Go 1.22+ and `git` on your PATH.

## Usage

```bash
gitto                              # Scan current directory for child repos
gitto ~/projects/repos             # Scan a specific directory
gitto --no-poll                    # Disable background status refresh
gitto -e vim                       # Override editor for this session
```

### Configuration

```bash
gitto config list                  # Show all settings
gitto config set editor "nvim"     # Set default editor
gitto config set poll_interval 10  # Poll every 10 seconds
gitto config set theme dark        # Force dark theme
```

Config is stored at `~/.config/gitto/config.json`.

| Key | Default | Description |
|-----|---------|-------------|
| `editor` | `$EDITOR` | Preferred editor |
| `theme` | `auto` | `auto`, `dark`, or `light` |
| `poll_interval` | `5` | Seconds between background refreshes |
| `show_untracked` | `true` | Show untracked files |
| `confirm_discard` | `true` | Prompt before discarding changes |
| `default_view` | `source` | `source`, `files`, or `history` |

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `s` | Source control view |
| `f` | File explorer view |
| `h` | History view (drill-in only) |
| `j/k` or `arrows` | Navigate |
| `R` | Force refresh all repos |
| `q` | Quit |
| `?` | Help overlay |

### Overview Mode

| Key | Action |
|-----|--------|
| `enter` | Expand/collapse repo |
| `o` | Open (drill into) repo |
| `space` | Stage/unstage file |
| `S` | Stage all in focused repo |
| `U` | Unstage all in focused repo |
| `c` | Start typing commit message |
| `p` | Push |
| `l` | Pull |
| `F` | Fetch |
| `b` | Branch picker |
| `z` | Stash dialog |
| `e` | Open file in editor |
| `d` | Discard file changes |
| `tab` | Jump to next repo |

### Drill-in Mode — Source Control

| Key | Action |
|-----|--------|
| `esc` | Back to overview |
| `space` | Stage/unstage file |
| `S` / `U` | Stage / unstage all |
| `c` | Commit message input |
| `alt+s` | Submit commit |
| `alt+a` | Amend last commit |
| `alt+p` | Commit and push |
| `p` | Push |
| `l` | Pull |
| `F` | Fetch |
| `b` | Branch picker |
| `z` | Stash dialog |
| `e` | Open file in editor |
| `d` / `D` | Discard file / discard all |
| `tab` | Cycle focus sections |

### Drill-in Mode — File Explorer

| Key | Action |
|-----|--------|
| `enter` | Open file / toggle directory |
| `l` / `h` | Expand / collapse directory |
| `space` | Stage/unstage file |
| `y` | Copy file path |
| `g` / `G` | Jump to top / bottom |

### Drill-in Mode — History

| Key | Action |
|-----|--------|
| `enter` | View commit details |
| `y` | Copy commit hash |
| `K` | Cherry-pick commit |
| `V` | Revert commit |
| `g` / `G` | Jump to top / bottom |

### iTerm2 Note

For `alt+` keybindings to work in iTerm2, go to **Settings > Profiles > Keys > General** and set **Left Option key** to **Esc+**.

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework (Elm architecture)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) — Text input, viewport, key bindings
- `os/exec` + `git` CLI — Uses your git config, SSH keys, GPG signing

## License

MIT
