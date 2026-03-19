package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jack-braga/gitto/internal/ui/styles"
)

// Tab names.
const (
	Source  = "source"
	Files   = "files"
	History = "history"
)

// Model is the view tab switcher component.
type Model struct {
	ActiveTab string
	Width     int
}

// New creates a new tabs model with the given default tab.
func New(defaultTab string) Model {
	return Model{ActiveTab: defaultTab}
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
	tabs := []struct {
		key  string
		name string
		id   string
	}{
		{"s", "source control", Source},
		{"f", "files", Files},
		{"h", "history", History},
	}

	var parts []string
	for _, tab := range tabs {
		label := fmt.Sprintf("[%s] %s", tab.key, tab.name)
		if tab.id == m.ActiveTab {
			parts = append(parts, styles.TabActiveStyle.Render(label))
		} else {
			parts = append(parts, styles.TabInactiveStyle.Render(label))
		}
	}

	left := strings.Join(parts, "    ")

	// Build the viewing label for the right side
	viewName := m.ActiveTab
	if viewName == Source {
		viewName = "source control"
	}
	right := styles.ViewLabelStyle.Render("viewing: " + viewName)

	// Calculate spacing
	leftLen := len([]rune(left)) // approximate
	rightLen := len([]rune(right))
	gap := m.Width - leftLen - rightLen - 4
	if gap < 2 {
		gap = 2
	}

	return " " + left + strings.Repeat(" ", gap) + right
}
