package overlays

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jack-braga/gitto/internal/ui/styles"
)

// ConfirmResult is sent when a confirmation dialog is answered.
type ConfirmResult struct {
	Action    string
	Confirmed bool
}

// ConfirmDialog is a yes/no confirmation overlay.
type ConfirmDialog struct {
	Message string
	Action  string
	Active  bool
	Width   int
}

// Show opens the confirmation dialog.
func (cd *ConfirmDialog) Show(message, action string) {
	cd.Active = true
	cd.Message = message
	cd.Action = action
}

// Hide closes the confirmation dialog.
func (cd *ConfirmDialog) Hide() {
	cd.Active = false
}

// Update handles input for the confirmation dialog.
func (cd ConfirmDialog) Update(msg tea.KeyMsg) (ConfirmDialog, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("y", "Y", "enter"))):
		action := cd.Action
		cd.Hide()
		return cd, func() tea.Msg {
			return ConfirmResult{Action: action, Confirmed: true}
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("n", "N", "esc"))):
		cd.Hide()
		return cd, nil
	}
	return cd, nil
}

// View renders the confirmation dialog.
func (cd ConfirmDialog) View() string {
	if !cd.Active {
		return ""
	}

	return fmt.Sprintf("\n  %s\n\n  %s\n",
		styles.DeletedStyle.Render(cd.Message),
		fmt.Sprintf("%s yes  %s no",
			styles.FooterKeyStyle.Render("y"),
			styles.FooterKeyStyle.Render("n"),
		),
	)
}
