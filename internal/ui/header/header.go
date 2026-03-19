package header

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jack-braga/gitto/internal/ui/styles"
)

// Model is the header bar component.
type Model struct {
	Width     int
	Path      string
	RepoName  string
	Branch    string
	Upstream  string
	Ahead     int
	Behind    int
	IsDrillIn bool
}

// New creates a new header model.
func New(path string) Model {
	return Model{Path: path}
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
	if m.IsDrillIn {
		return m.drillInView()
	}
	return m.overviewView()
}

func (m Model) overviewView() string {
	title := styles.HeaderStyle.Render("gitto")
	path := styles.HeaderPathStyle.Render(m.Path)
	return fmt.Sprintf("%s %s", title, path)
}

func (m Model) drillInView() string {
	title := styles.HeaderStyle.Render("gitto")
	arrow := styles.HeaderPathStyle.Render(">")
	repo := styles.RepoNameFocusedStyle.Render(m.RepoName)
	branch := styles.BranchPill.Render(m.Branch)

	line := fmt.Sprintf("%s %s %s  %s", title, arrow, repo, branch)

	if m.Upstream != "" {
		line += "  " + styles.HeaderPathStyle.Render(m.Upstream)
	}

	if m.Ahead > 0 {
		line += "  " + styles.AheadStyle.Render(fmt.Sprintf("↑%d", m.Ahead))
	}
	if m.Behind > 0 {
		line += " " + styles.BehindStyle.Render(fmt.Sprintf("↓%d", m.Behind))
	}

	return line
}
