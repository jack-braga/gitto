package overlays

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	t "github.com/jack-braga/gitto/internal/types"
	"github.com/jack-braga/gitto/internal/git"
	"github.com/jack-braga/gitto/internal/ui/styles"
)

// StashDialogMode tracks what the stash dialog is doing.
type StashDialogMode int

const (
	StashBrowse StashDialogMode = iota
	StashSaving
)

// StashDialog is the stash management overlay.
type StashDialog struct {
	Stashes  []git.Stash
	Cursor   int
	RepoPath string
	Width    int
	Mode     StashDialogMode
	Input    textinput.Model
	Active   bool
}

// NewStashDialog creates a new stash dialog.
func NewStashDialog() StashDialog {
	ti := textinput.New()
	ti.Placeholder = "Stash message (optional)..."
	ti.CharLimit = 200
	return StashDialog{Input: ti}
}

// Show opens the stash dialog.
func (sd *StashDialog) Show(repoPath string, stashes []git.Stash) {
	sd.Active = true
	sd.RepoPath = repoPath
	sd.Stashes = stashes
	sd.Cursor = 0
	sd.Mode = StashBrowse
}

// Hide closes the stash dialog.
func (sd *StashDialog) Hide() {
	sd.Active = false
	sd.Input.Blur()
}

// Update handles input for the stash dialog.
func (sd StashDialog) Update(msg tea.KeyMsg) (StashDialog, tea.Cmd) {
	if sd.Mode == StashSaving {
		return sd.updateSave(msg)
	}

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		sd.Hide()
	case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
		if sd.Cursor < len(sd.Stashes)-1 {
			sd.Cursor++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
		if sd.Cursor > 0 {
			sd.Cursor--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
		sd.Mode = StashSaving
		sd.Input.SetValue("")
		sd.Input.Focus()
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		// Apply stash
		if sd.Cursor < len(sd.Stashes) {
			idx := sd.Stashes[sd.Cursor].Index
			repoPath := sd.RepoPath
			return sd, func() tea.Msg {
				err := git.StashApply(repoPath, idx)
				return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "stash-apply", Err: err}
			}
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("p"))):
		// Pop stash
		if sd.Cursor < len(sd.Stashes) {
			idx := sd.Stashes[sd.Cursor].Index
			repoPath := sd.RepoPath
			sd.Hide()
			return sd, func() tea.Msg {
				err := git.StashPop(repoPath, idx)
				return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "stash-pop", Err: err}
			}
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("D"))):
		// Drop stash
		if sd.Cursor < len(sd.Stashes) {
			idx := sd.Stashes[sd.Cursor].Index
			repoPath := sd.RepoPath
			return sd, func() tea.Msg {
				err := git.StashDrop(repoPath, idx)
				return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "stash-drop", Err: err}
			}
		}
	}

	return sd, nil
}

func (sd StashDialog) updateSave(msg tea.KeyMsg) (StashDialog, tea.Cmd) {
	switch msg.String() {
	case "esc":
		sd.Mode = StashBrowse
		sd.Input.Blur()
		return sd, nil
	case "enter":
		message := sd.Input.Value()
		repoPath := sd.RepoPath
		sd.Hide()
		return sd, func() tea.Msg {
			err := git.StashSave(repoPath, message)
			return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "stash-save", Err: err}
		}
	default:
		var cmd tea.Cmd
		sd.Input, cmd = sd.Input.Update(msg)
		return sd, cmd
	}
}

// View renders the stash dialog.
func (sd StashDialog) View() string {
	if !sd.Active {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.HeaderStyle.Render("Stashes") + "\n")
	b.WriteString(styles.Divider(sd.Width-4) + "\n")

	if sd.Mode == StashSaving {
		b.WriteString("  Save stash: " + sd.Input.View() + "\n")
		b.WriteString(styles.Divider(sd.Width-4) + "\n")
	}

	if len(sd.Stashes) == 0 {
		b.WriteString("  " + styles.FooterStyle.Render("No stashes") + "\n")
	} else {
		for i, s := range sd.Stashes {
			focused := i == sd.Cursor
			line := fmt.Sprintf("  stash@{%d}  %s  %s",
				s.Index,
				s.Message,
				styles.DateStyle.Render(styles.RelativeTime(s.Date)),
			)
			if focused {
				b.WriteString(styles.SelectedStyle.Render(line) + "\n")
			} else {
				b.WriteString(line + "\n")
			}
		}
	}

	b.WriteString("\n")
	hints := fmt.Sprintf("  %s save  %s apply  %s pop  %s drop  %s close",
		styles.FooterKeyStyle.Render("s"),
		styles.FooterKeyStyle.Render("enter"),
		styles.FooterKeyStyle.Render("p"),
		styles.FooterKeyStyle.Render("D"),
		styles.FooterKeyStyle.Render("esc"),
	)
	b.WriteString(hints + "\n")

	return b.String()
}
