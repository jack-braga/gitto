# gitto

A multi-repo Git TUI workspace manager built with Go + Bubble Tea.

Run `gitto` in any parent directory to manage all child Git repos from one terminal interface.

## Install

```bash
go install github.com/jack-braga/gitto@latest
```

## Usage

```bash
gitto                           # Run in current directory
gitto ~/projects/repos          # Run in a specific directory
gitto config set editor "nvim"  # Set default editor
```

## Keybindings

### Overview Mode

| Key | Action |
|-----|--------|
| `enter` | Expand/collapse repo |
| `o` | Open (drill into) repo |
| `space` | Stage/unstage file |
| `c` | Commit message input |
| `p` | Push |
| `l` | Pull |
| `b` | Branch picker |
| `j/k` | Navigate |
| `?` | Help |

### Drill-in Mode

| Key | Action |
|-----|--------|
| `esc` | Back to overview |
| `s/f/h` | Source control / files / history |
| `alt+s` | Submit commit |
| `alt+a` | Amend last commit |
| `alt+p` | Commit and push |

## License

MIT
