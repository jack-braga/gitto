package footer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jack-braga/gitto/internal/ui/styles"
)

// Model is the context-sensitive footer with keybinding hints.
type Model struct {
	Width        int
	Mode         string // "overview" or "drillin"
	ActiveView   string // "source", "files", "history"
	IsCommitting bool
	HasOverlay   bool
}

// New creates a new footer model.
func New() Model {
	return Model{Mode: "overview", ActiveView: "source"}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.Width = msg.Width
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.HasOverlay {
		return ""
	}

	var hints [][]string // pairs of [key, description]

	switch {
	case m.Mode == "overview" && m.ActiveView =="source":
		hints = [][]string{
			{"enter", "expand/collapse"},
			{"o", "open repo"},
			{"space", "stage/unstage"},
			{"c", "commit"},
			{"p", "push"},
			{"l", "pull"},
			{"b", "branch"},
			{"?", "help"},
		}
	case m.Mode == "drillin" && m.ActiveView =="source":
		hints = [][]string{
			{"esc", "back"},
			{"space", "stage/unstage"},
			{"S", "stage all"},
			{"e", "open in editor"},
			{"p", "push"},
			{"l", "pull"},
			{"F", "fetch"},
			{"b", "branches"},
			{"z", "stash"},
			{"?", "help"},
		}
	case m.Mode == "drillin" && m.ActiveView =="files":
		hints = [][]string{
			{"enter", "open"},
			{"space", "stage/unstage"},
			{"l/h", "expand/collapse"},
			{"j/k", "navigate"},
			{"y", "copy path"},
			{"esc", "back"},
		}
	case m.Mode == "drillin" && m.ActiveView =="history":
		hints = [][]string{
			{"enter", "view details"},
			{"y", "copy hash"},
			{"K", "cherry-pick"},
			{"V", "revert"},
			{"esc", "back"},
			{"j/k", "navigate"},
		}
	default:
		hints = [][]string{
			{"j/k", "navigate"},
			{"?", "help"},
		}
	}

	return renderHints(hints)
}

func renderHints(hints [][]string) string {
	var parts []string
	for _, h := range hints {
		key := styles.FooterKeyStyle.Render(h[0])
		desc := styles.FooterStyle.Render(" " + h[1])
		parts = append(parts, key+desc)
	}
	return styles.FooterStyle.Render("  " + strings.Join(parts, "  "))
}
