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

// BranchPickerMode tracks what the branch picker is doing.
type BranchPickerMode int

const (
	BranchBrowse BranchPickerMode = iota
	BranchCreate
	BranchFilter
)

// BranchPicker is the branch list overlay.
type BranchPicker struct {
	Branches   []git.Branch
	Filtered   []git.Branch
	Cursor     int
	RepoPath   string
	Width      int
	Height     int
	Mode       BranchPickerMode
	Input      textinput.Model
	FilterText string
	Active     bool
}

// NewBranchPicker creates a new branch picker.
func NewBranchPicker() BranchPicker {
	ti := textinput.New()
	ti.Placeholder = "Branch name..."
	ti.CharLimit = 200
	return BranchPicker{Input: ti}
}

// Show opens the branch picker for a repo.
func (bp *BranchPicker) Show(repoPath string, branches []git.Branch) {
	bp.Active = true
	bp.RepoPath = repoPath
	bp.Branches = branches
	bp.Filtered = branches
	bp.Cursor = 0
	bp.Mode = BranchBrowse
	bp.FilterText = ""
}

// Hide closes the branch picker.
func (bp *BranchPicker) Hide() {
	bp.Active = false
	bp.Input.Blur()
}

// Update handles input for the branch picker.
func (bp BranchPicker) Update(msg tea.KeyMsg) (BranchPicker, tea.Cmd) {
	if bp.Mode == BranchCreate {
		return bp.updateCreate(msg)
	}
	if bp.Mode == BranchFilter {
		return bp.updateFilter(msg)
	}

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		bp.Hide()
	case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
		if bp.Cursor < len(bp.Filtered)-1 {
			bp.Cursor++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
		if bp.Cursor > 0 {
			bp.Cursor--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		if bp.Cursor < len(bp.Filtered) {
			branch := bp.Filtered[bp.Cursor]
			repoPath := bp.RepoPath
			bp.Hide()
			return bp, func() tea.Msg {
				err := git.SwitchBranch(repoPath, branch.Name)
				return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "checkout", Err: err}
			}
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
		bp.Mode = BranchCreate
		bp.Input.SetValue("")
		bp.Input.Focus()
	case key.Matches(msg, key.NewBinding(key.WithKeys("D"))):
		if bp.Cursor < len(bp.Filtered) {
			branch := bp.Filtered[bp.Cursor]
			if !branch.IsCurrent && !branch.IsRemote {
				repoPath := bp.RepoPath
				return bp, func() tea.Msg {
					err := git.DeleteBranch(repoPath, branch.Name)
					return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "delete-branch", Err: err}
				}
			}
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
		bp.Mode = BranchFilter
		bp.Input.SetValue("")
		bp.Input.Placeholder = "Filter branches..."
		bp.Input.Focus()
	}

	return bp, nil
}

func (bp BranchPicker) updateCreate(msg tea.KeyMsg) (BranchPicker, tea.Cmd) {
	switch msg.String() {
	case "esc":
		bp.Mode = BranchBrowse
		bp.Input.Blur()
		return bp, nil
	case "enter":
		name := bp.Input.Value()
		if name == "" {
			return bp, nil
		}
		repoPath := bp.RepoPath
		bp.Hide()
		return bp, func() tea.Msg {
			err := git.CreateBranch(repoPath, name)
			return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "create-branch", Err: err}
		}
	default:
		var cmd tea.Cmd
		bp.Input, cmd = bp.Input.Update(msg)
		return bp, cmd
	}
}

func (bp BranchPicker) updateFilter(msg tea.KeyMsg) (BranchPicker, tea.Cmd) {
	switch msg.String() {
	case "esc":
		bp.Mode = BranchBrowse
		bp.Input.Blur()
		bp.Filtered = bp.Branches
		bp.Cursor = 0
		return bp, nil
	case "enter":
		bp.Mode = BranchBrowse
		bp.Input.Blur()
		return bp, nil
	default:
		var cmd tea.Cmd
		bp.Input, cmd = bp.Input.Update(msg)
		// Apply filter
		bp.FilterText = bp.Input.Value()
		bp.Filtered = nil
		for _, b := range bp.Branches {
			if strings.Contains(strings.ToLower(b.Name), strings.ToLower(bp.FilterText)) {
				bp.Filtered = append(bp.Filtered, b)
			}
		}
		bp.Cursor = 0
		return bp, cmd
	}
}

// View renders the branch picker overlay.
func (bp BranchPicker) View() string {
	if !bp.Active {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.HeaderStyle.Render("Branches") + "\n")
	b.WriteString(styles.Divider(bp.Width-4) + "\n")

	if bp.Mode == BranchCreate {
		b.WriteString("  New branch: " + bp.Input.View() + "\n")
		b.WriteString(styles.Divider(bp.Width-4) + "\n")
	}

	if bp.Mode == BranchFilter {
		b.WriteString("  Filter: " + bp.Input.View() + "\n")
		b.WriteString(styles.Divider(bp.Width-4) + "\n")
	}

	for i, branch := range bp.Filtered {
		focused := i == bp.Cursor

		name := branch.Name
		if branch.IsCurrent {
			name = "* " + name
		}
		if branch.IsRemote {
			name = styles.UntrackedStyle.Render(name)
		}

		line := "  " + name
		if branch.LastMessage != "" {
			line += "  " + styles.FooterStyle.Render(styles.Truncate(branch.LastMessage, 40))
		}

		if focused {
			b.WriteString(styles.SelectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}

	b.WriteString("\n")
	hints := fmt.Sprintf("  %s switch  %s new  %s delete  %s filter  %s close",
		styles.FooterKeyStyle.Render("enter"),
		styles.FooterKeyStyle.Render("n"),
		styles.FooterKeyStyle.Render("D"),
		styles.FooterKeyStyle.Render("/"),
		styles.FooterKeyStyle.Render("esc"),
	)
	b.WriteString(hints + "\n")

	return b.String()
}
