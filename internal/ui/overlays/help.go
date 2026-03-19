package overlays

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jack-braga/gitto/internal/ui/styles"
)

// HelpOverlay shows the full keybinding reference.
type HelpOverlay struct {
	Active bool
	Width  int
}

// Toggle flips the help overlay on/off.
func (h *HelpOverlay) Toggle() {
	h.Active = !h.Active
}

// Update handles input for the help overlay.
func (h HelpOverlay) Update(msg tea.KeyMsg) (HelpOverlay, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("?", "esc", "q"))) {
		h.Active = false
	}
	return h, nil
}

// View renders the help overlay.
func (h HelpOverlay) View() string {
	if !h.Active {
		return ""
	}

	sections := []struct {
		title string
		keys  [][]string
	}{
		{
			"Global",
			[][]string{
				{"s", "Source control view"},
				{"f", "File explorer view"},
				{"h", "History view"},
				{"j/k", "Navigate up/down"},
				{"R", "Force refresh all"},
				{"q", "Quit"},
				{"?", "Toggle help"},
			},
		},
		{
			"Overview Mode",
			[][]string{
				{"enter", "Drill into repo"},
				{"space", "Stage/unstage file"},
				{"S", "Stage all in repo"},
				{"U", "Unstage all in repo"},
				{"c", "Commit message input"},
				{"p", "Push"},
				{"l", "Pull"},
				{"F", "Fetch"},
				{"b", "Branch picker"},
				{"z", "Stash dialog"},
				{"e", "Open file in editor"},
				{"d", "Discard file changes"},
				{"tab", "Next repo"},
			},
		},
		{
			"Drill-in Source Control",
			[][]string{
				{"esc", "Back to overview"},
				{"space", "Stage/unstage file"},
				{"S", "Stage all"},
				{"U", "Unstage all"},
				{"c", "Commit message input"},
				{"ctrl+s", "Submit commit"},
				{"ctrl+a", "Amend last commit"},
				{"ctrl+p", "Commit and push"},
				{"p", "Push"},
				{"l", "Pull"},
				{"F", "Fetch"},
				{"b", "Branch picker"},
				{"z", "Stash dialog"},
				{"e", "Open in editor"},
				{"d", "Discard file"},
				{"D", "Discard all"},
				{"tab", "Cycle focus sections"},
			},
		},
		{
			"File Explorer",
			[][]string{
				{"enter", "Open file / toggle dir"},
				{"l", "Expand directory"},
				{"h", "Collapse directory"},
				{"space", "Stage/unstage file"},
				{"y", "Copy file path"},
				{"g", "Jump to top"},
				{"G", "Jump to bottom"},
			},
		},
		{
			"History",
			[][]string{
				{"enter", "View commit details"},
				{"y", "Copy commit hash"},
				{"K", "Cherry-pick commit"},
				{"V", "Revert commit"},
				{"g", "Jump to top"},
				{"G", "Jump to bottom"},
			},
		},
		{
			"Branch Picker",
			[][]string{
				{"enter", "Switch to branch"},
				{"n", "Create new branch"},
				{"D", "Delete branch"},
				{"/", "Filter branches"},
				{"esc", "Close"},
			},
		},
		{
			"Stash Dialog",
			[][]string{
				{"s", "Save new stash"},
				{"enter", "Apply stash"},
				{"p", "Pop stash"},
				{"D", "Drop stash"},
				{"esc", "Close"},
			},
		},
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.HeaderStyle.Render("Keybindings") + "\n")
	b.WriteString(styles.Divider(h.Width-4) + "\n\n")

	for _, section := range sections {
		b.WriteString("  " + styles.CommitLabelStyle.Render(section.title) + "\n")
		for _, kv := range section.keys {
			k := styles.FooterKeyStyle.Render(kv[0])
			padded := kv[0]
			gap := 14 - len(padded)
			if gap < 1 {
				gap = 1
			}
			b.WriteString("    " + k + strings.Repeat(" ", gap) + styles.FooterStyle.Render(kv[1]) + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("  " + styles.FooterStyle.Render("Press ? or esc to close") + "\n")

	return b.String()
}
