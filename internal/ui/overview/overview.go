package overview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jack-braga/gitto/internal/git"
	t "github.com/jack-braga/gitto/internal/types"
	"github.com/jack-braga/gitto/internal/ui/styles"
)

// FocusSection tracks what area the user is focused on.
type FocusSection int

const (
	FocusRepoList FocusSection = iota
	FocusFiles
	FocusCommitInput
)

// Model is the overview mode root model showing all repos.
type Model struct {
	Repos       []*git.Repository
	Width       int
	Height      int
	FocusedRepo int // Index into Repos
	FocusedFile int // Index into the files of the focused repo
	Focus       FocusSection
	Expanded    map[int]bool // Which repos are expanded
	CommitInputs map[int]textinput.Model // Per-repo commit inputs
	Keys        t.KeyMap
	StatusMsg   string // Transient status message
}

// New creates a new overview model.
func New(keys t.KeyMap) Model {
	return Model{
		Expanded:     make(map[int]bool),
		CommitInputs: make(map[int]textinput.Model),
		Keys:         keys,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetRepos updates the repos list.
func (m *Model) SetRepos(repos []*git.Repository) {
	m.Repos = repos
	// Auto-expand repos with changes only if not already tracked
	for i, r := range repos {
		if _, exists := m.Expanded[i]; !exists {
			if !r.IsClean() {
				m.Expanded[i] = true
			}
		}
	}
}

// UpdateRepoStatus updates a single repo's status.
func (m *Model) UpdateRepoStatus(path string, status *git.Repository) {
	for i, r := range m.Repos {
		if r.Path == path {
			m.Repos[i] = status
			return
		}
	}
}

// FocusedRepository returns the currently focused repo, or nil.
func (m Model) FocusedRepository() *git.Repository {
	if m.FocusedRepo >= 0 && m.FocusedRepo < len(m.Repos) {
		return m.Repos[m.FocusedRepo]
	}
	return nil
}

// FocusedFileChange returns the currently focused file, its type, or nil.
func (m Model) FocusedFileChange() (*git.FileChange, string) {
	repo := m.FocusedRepository()
	if repo == nil || m.Focus != FocusFiles {
		return nil, ""
	}

	idx := m.FocusedFile
	if idx < len(repo.Staged) {
		return &repo.Staged[idx], "staged"
	}
	idx -= len(repo.Staged)
	if idx < len(repo.Unstaged) {
		return &repo.Unstaged[idx], "unstaged"
	}
	idx -= len(repo.Unstaged)
	if idx < len(repo.Untracked) {
		return &repo.Untracked[idx], "untracked"
	}
	return nil, ""
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case tea.KeyMsg:
		// If we're in commit input mode, handle that first
		if m.Focus == FocusCommitInput {
			return m.updateCommitInput(msg)
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.Keys.Down):
		m.moveDown()
	case key.Matches(msg, m.Keys.Up):
		m.moveUp()
	case key.Matches(msg, m.Keys.Tab):
		m.nextRepo()
	case key.Matches(msg, m.Keys.ShiftTab):
		m.prevRepo()
	case key.Matches(msg, m.Keys.Space):
		return m.toggleStage()
	case key.Matches(msg, m.Keys.StageAll):
		return m.stageAll()
	case key.Matches(msg, m.Keys.UnstageAll):
		return m.unstageAll()
	case key.Matches(msg, m.Keys.Commit):
		m.enterCommitInput()
	case key.Matches(msg, m.Keys.Push):
		return m.push()
	case key.Matches(msg, m.Keys.Pull):
		return m.pull()
	case key.Matches(msg, m.Keys.FetchKey):
		return m.fetch()
	case key.Matches(msg, m.Keys.Discard):
		return m.discard()
	case key.Matches(msg, m.Keys.Enter):
		// Toggle expand/collapse
		if m.Focus == FocusRepoList {
			m.Expanded[m.FocusedRepo] = !m.Expanded[m.FocusedRepo]
			if !m.Expanded[m.FocusedRepo] {
				// Collapsing — reset file focus
				m.Focus = FocusRepoList
			}
		}
	case key.Matches(msg, m.Keys.JumpTop):
		m.FocusedRepo = 0
		m.FocusedFile = 0
		m.Focus = FocusRepoList
	case key.Matches(msg, m.Keys.JumpBottom):
		if len(m.Repos) > 0 {
			m.FocusedRepo = len(m.Repos) - 1
			m.Focus = FocusRepoList
		}
	case key.Matches(msg, m.Keys.HalfPageDown):
		for i := 0; i < m.Height/2; i++ {
			m.moveDown()
		}
	case key.Matches(msg, m.Keys.HalfPageUp):
		for i := 0; i < m.Height/2; i++ {
			m.moveUp()
		}
	case key.Matches(msg, m.Keys.DrillIn):
		// Drill into the focused repo
		if m.Focus == FocusRepoList || m.Focus == FocusFiles {
			return m, func() tea.Msg {
				return t.DrillInMsg{RepoIndex: m.FocusedRepo}
			}
		}
	}
	return m, nil
}

func (m *Model) moveDown() {
	if m.Focus == FocusRepoList {
		if m.Expanded[m.FocusedRepo] && m.FocusedRepository() != nil && !m.FocusedRepository().IsClean() {
			// Move into files
			m.Focus = FocusFiles
			m.FocusedFile = 0
		} else if m.FocusedRepo < len(m.Repos)-1 {
			m.FocusedRepo++
		}
	} else if m.Focus == FocusFiles {
		repo := m.FocusedRepository()
		if repo == nil {
			return
		}
		totalFiles := len(repo.Staged) + len(repo.Unstaged) + len(repo.Untracked)
		if m.FocusedFile < totalFiles-1 {
			m.FocusedFile++
		} else if m.FocusedRepo < len(m.Repos)-1 {
			// Move to next repo
			m.FocusedRepo++
			m.Focus = FocusRepoList
		}
	}
}

func (m *Model) moveUp() {
	if m.Focus == FocusFiles {
		if m.FocusedFile > 0 {
			m.FocusedFile--
		} else {
			m.Focus = FocusRepoList
		}
	} else if m.Focus == FocusRepoList {
		if m.FocusedRepo > 0 {
			m.FocusedRepo--
			// If previous repo is expanded with files, focus last file
			repo := m.FocusedRepository()
			if m.Expanded[m.FocusedRepo] && repo != nil && !repo.IsClean() {
				m.Focus = FocusFiles
				m.FocusedFile = len(repo.Staged) + len(repo.Unstaged) + len(repo.Untracked) - 1
			}
		}
	}
}

func (m *Model) nextRepo() {
	if m.FocusedRepo < len(m.Repos)-1 {
		m.FocusedRepo++
		m.Focus = FocusRepoList
		m.FocusedFile = 0
	}
}

func (m *Model) prevRepo() {
	if m.FocusedRepo > 0 {
		m.FocusedRepo--
		m.Focus = FocusRepoList
		m.FocusedFile = 0
	}
}

func (m Model) toggleStage() (Model, tea.Cmd) {
	fc, section := m.FocusedFileChange()
	if fc == nil {
		return m, nil
	}
	repo := m.FocusedRepository()

	switch section {
	case "staged":
		return m, func() tea.Msg {
			err := git.UnstageFile(repo.Path, fc.Path)
			return t.GitOpCompleteMsg{RepoPath: repo.Path, Op: "unstage", Err: err}
		}
	case "unstaged", "untracked":
		return m, func() tea.Msg {
			err := git.StageFile(repo.Path, fc.Path)
			return t.GitOpCompleteMsg{RepoPath: repo.Path, Op: "stage", Err: err}
		}
	}
	return m, nil
}

func (m Model) stageAll() (Model, tea.Cmd) {
	repo := m.FocusedRepository()
	if repo == nil {
		return m, nil
	}
	return m, func() tea.Msg {
		err := git.StageAll(repo.Path)
		return t.GitOpCompleteMsg{RepoPath: repo.Path, Op: "stage", Err: err}
	}
}

func (m Model) unstageAll() (Model, tea.Cmd) {
	repo := m.FocusedRepository()
	if repo == nil {
		return m, nil
	}
	return m, func() tea.Msg {
		err := git.UnstageAll(repo.Path)
		return t.GitOpCompleteMsg{RepoPath: repo.Path, Op: "unstage", Err: err}
	}
}

func (m *Model) enterCommitInput() {
	repo := m.FocusedRepository()
	if repo == nil {
		return
	}
	if _, ok := m.CommitInputs[m.FocusedRepo]; !ok {
		ti := textinput.New()
		ti.Placeholder = "Commit message..."
		ti.CharLimit = 500
		ti.Width = m.Width - 10
		m.CommitInputs[m.FocusedRepo] = ti
	}
	ti := m.CommitInputs[m.FocusedRepo]
	ti.Focus()
	m.CommitInputs[m.FocusedRepo] = ti
	m.Focus = FocusCommitInput
}

func (m Model) updateCommitInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		ti := m.CommitInputs[m.FocusedRepo]
		ti.Blur()
		m.CommitInputs[m.FocusedRepo] = ti
		m.Focus = FocusRepoList
		return m, nil
	case "alt+s":
		return m.submitCommit(false)
	case "alt+p":
		return m.submitCommit(true)
	default:
		ti := m.CommitInputs[m.FocusedRepo]
		var cmd tea.Cmd
		ti, cmd = ti.Update(msg)
		m.CommitInputs[m.FocusedRepo] = ti
		return m, cmd
	}
}

func (m Model) submitCommit(andPush bool) (Model, tea.Cmd) {
	ti := m.CommitInputs[m.FocusedRepo]
	message := ti.Value()
	if message == "" {
		return m, nil
	}
	repo := m.FocusedRepository()
	if repo == nil {
		return m, nil
	}

	ti.SetValue("")
	ti.Blur()
	m.CommitInputs[m.FocusedRepo] = ti
	m.Focus = FocusRepoList

	return m, func() tea.Msg {
		err := git.Commit(repo.Path, message)
		if err != nil {
			return t.GitOpCompleteMsg{RepoPath: repo.Path, Op: "commit", Err: err}
		}
		if andPush {
			output, pushErr := git.Push(repo.Path)
			return t.GitOpCompleteMsg{RepoPath: repo.Path, Op: "commit+push", Err: pushErr, Output: output}
		}
		return t.GitOpCompleteMsg{RepoPath: repo.Path, Op: "commit", Err: nil}
	}
}

func (m Model) push() (Model, tea.Cmd) {
	repo := m.FocusedRepository()
	if repo == nil {
		return m, nil
	}
	return m, func() tea.Msg {
		output, err := git.Push(repo.Path)
		return t.GitOpCompleteMsg{RepoPath: repo.Path, Op: "push", Err: err, Output: output}
	}
}

func (m Model) pull() (Model, tea.Cmd) {
	repo := m.FocusedRepository()
	if repo == nil {
		return m, nil
	}
	return m, func() tea.Msg {
		output, err := git.Pull(repo.Path)
		return t.GitOpCompleteMsg{RepoPath: repo.Path, Op: "pull", Err: err, Output: output}
	}
}

func (m Model) fetch() (Model, tea.Cmd) {
	repo := m.FocusedRepository()
	if repo == nil {
		return m, nil
	}
	return m, func() tea.Msg {
		output, err := git.Fetch(repo.Path)
		return t.GitOpCompleteMsg{RepoPath: repo.Path, Op: "fetch", Err: err, Output: output}
	}
}

func (m Model) discard() (Model, tea.Cmd) {
	fc, section := m.FocusedFileChange()
	if fc == nil {
		return m, nil
	}
	repo := m.FocusedRepository()
	isUntracked := section == "untracked"

	return m, func() tea.Msg {
		err := git.DiscardFile(repo.Path, fc.Path, isUntracked)
		return t.GitOpCompleteMsg{RepoPath: repo.Path, Op: "discard", Err: err}
	}
}

// CursorLine returns the line number (0-based) of the focused item in the rendered output.
func (m Model) CursorLine() int {
	line := 0
	for i, repo := range m.Repos {
		// Repo header line
		if i == m.FocusedRepo && m.Focus == FocusRepoList {
			return line
		}
		line++ // repo header
		// Expanded content
		if m.Expanded[i] {
			if len(repo.Staged) > 0 {
				line++ // "Staged (N)" header
				for j := range repo.Staged {
					if i == m.FocusedRepo && m.Focus == FocusFiles && m.FocusedFile == j {
						return line
					}
					line++
				}
			}
			if len(repo.Unstaged) > 0 {
				line++ // "Changes (N)" header
				for j := range repo.Unstaged {
					globalIdx := len(repo.Staged) + j
					if i == m.FocusedRepo && m.Focus == FocusFiles && m.FocusedFile == globalIdx {
						return line
					}
					line++
				}
			}
			if len(repo.Untracked) > 0 {
				line++ // "Untracked (N)" header
				for j := range repo.Untracked {
					globalIdx := len(repo.Staged) + len(repo.Unstaged) + j
					if i == m.FocusedRepo && m.Focus == FocusFiles && m.FocusedFile == globalIdx {
						return line
					}
					line++
				}
			}
			// Commit input (3 lines)
			if _, ok := m.CommitInputs[i]; ok && m.Focus == FocusCommitInput && i == m.FocusedRepo {
				line += 3
			}
		}
		line++ // blank line after repo
	}
	return line
}

// View implements tea.Model.
func (m Model) View() string {
	if len(m.Repos) == 0 {
		return "\n  No git repositories found in this directory.\n"
	}

	var b strings.Builder

	for i, repo := range m.Repos {
		isFocused := i == m.FocusedRepo && m.Focus == FocusRepoList

		// Repo header line
		b.WriteString(m.renderRepoHeader(i, repo, isFocused))
		b.WriteString("\n")

		// Expanded content
		if m.Expanded[i] {
			b.WriteString(m.renderRepoFiles(i, repo))

			// Commit input
			if ti, ok := m.CommitInputs[i]; ok && m.Focus == FocusCommitInput && i == m.FocusedRepo {
				b.WriteString("  " + styles.CommitLabelStyle.Render("COMMIT MESSAGE") + "\n")
				b.WriteString("  " + styles.CommitInputStyle.Render(ti.View()) + "\n")
				b.WriteString("  " + styles.FooterStyle.Render("alt+s submit  alt+p commit+push  esc cancel") + "\n")
			}
		}
		b.WriteString("\n")
	}

	if m.StatusMsg != "" {
		b.WriteString("\n  " + m.StatusMsg + "\n")
	}

	return b.String()
}

func (m Model) renderRepoHeader(idx int, repo *git.Repository, focused bool) string {
	cursor := styles.Cursor(focused)

	// Expand/collapse icon
	icon := styles.CollapsedIcon
	if m.Expanded[idx] {
		icon = styles.ExpandedIcon
	}

	// Repo name
	name := repo.Name
	if focused {
		name = styles.RepoNameFocusedStyle.Render(name)
	} else {
		name = styles.RepoNameStyle.Render(name)
	}

	// Branch pill
	branch := styles.BranchPill.Render(repo.Branch)

	// Change summary
	var changeParts []string
	if n := len(repo.Staged) + len(repo.Unstaged); n > 0 {
		changeParts = append(changeParts, styles.ModifiedStyle.Render(fmt.Sprintf("%dM", n)))
	}
	if n := len(repo.Untracked); n > 0 {
		changeParts = append(changeParts, styles.AddedStyle.Render(fmt.Sprintf("%dA", n)))
	}
	if repo.IsClean() {
		changeParts = append(changeParts, styles.CleanStyle.Render("clean"))
	}

	// Ahead/behind
	var abParts []string
	if repo.Ahead > 0 {
		abParts = append(abParts, styles.AheadStyle.Render(fmt.Sprintf("↑%d", repo.Ahead)))
	}
	if repo.Behind > 0 {
		abParts = append(abParts, styles.BehindStyle.Render(fmt.Sprintf("↓%d", repo.Behind)))
	}

	line := fmt.Sprintf("%s%s %s  %s  %s", cursor, icon, name, branch, strings.Join(changeParts, " "))
	if len(abParts) > 0 {
		line += "  " + strings.Join(abParts, " ")
	}

	if focused {
		return styles.SelectedStyle.Render(line)
	}
	return line
}

func (m Model) renderRepoFiles(idx int, repo *git.Repository) string {
	var b strings.Builder

	if len(repo.Staged) > 0 {
		b.WriteString("  " + styles.StagedHeaderStyle.Render(fmt.Sprintf("Staged (%d)", len(repo.Staged))) + "\n")
		for j, f := range repo.Staged {
			globalIdx := j
			focused := idx == m.FocusedRepo && m.Focus == FocusFiles && m.FocusedFile == globalIdx
			b.WriteString(m.renderFileRow(f, focused))
		}
	}

	if len(repo.Unstaged) > 0 {
		b.WriteString("  " + styles.UnstagedHeaderStyle.Render(fmt.Sprintf("Changes (%d)", len(repo.Unstaged))) + "\n")
		for j, f := range repo.Unstaged {
			globalIdx := len(repo.Staged) + j
			focused := idx == m.FocusedRepo && m.Focus == FocusFiles && m.FocusedFile == globalIdx
			b.WriteString(m.renderFileRow(f, focused))
		}
	}

	if len(repo.Untracked) > 0 {
		b.WriteString("  " + styles.UnstagedHeaderStyle.Render(fmt.Sprintf("Untracked (%d)", len(repo.Untracked))) + "\n")
		for j, f := range repo.Untracked {
			globalIdx := len(repo.Staged) + len(repo.Unstaged) + j
			focused := idx == m.FocusedRepo && m.Focus == FocusFiles && m.FocusedFile == globalIdx
			b.WriteString(m.renderFileRow(f, focused))
		}
	}

	return b.String()
}

func (m Model) renderFileRow(f git.FileChange, focused bool) string {
	cursor := styles.Cursor(focused)
	status := styles.StatusStyle(f.Status).Render(styles.StatusChar(f.Status))
	line := fmt.Sprintf("  %s%s  %s", cursor, status, f.Path)
	if focused {
		return styles.SelectedStyle.Render(line) + "\n"
	}
	return line + "\n"
}
