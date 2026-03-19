package drillin

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	t "github.com/jack-braga/gitto/internal/types"
	"github.com/jack-braga/gitto/internal/config"
	"github.com/jack-braga/gitto/internal/editor"
	"github.com/jack-braga/gitto/internal/git"
	"github.com/jack-braga/gitto/internal/ui/styles"
)

// FocusSection tracks what area of drill-in the user is focused on.
type FocusSection int

const (
	FocusCommit FocusSection = iota
	FocusStaged
	FocusUnstaged
	FocusStashes
)

// ActiveView tracks which view is active.
const (
	ViewSource  = "source"
	ViewFiles   = "files"
	ViewHistory = "history"
)

// Model is the drill-in mode for a single repository.
type Model struct {
	Repo        *git.Repository
	Width       int
	Height      int
	Keys        t.KeyMap
	Config      *config.Config
	ActiveView  string

	// Source control
	Focus       FocusSection
	CommitInput textinput.Model
	FocusedFile int
	Stashes     []git.Stash

	// File explorer
	FileTree    *git.FileTreeNode
	FlatTree    []*git.FileTreeNode // Flattened visible nodes
	TreeCursor  int

	// History
	LogEntries  []git.LogEntry
	LogCursor   int

	StatusMsg   string
}

// New creates a new drill-in model for the given repo.
func New(repo *git.Repository, keys t.KeyMap, cfg *config.Config, width, height int) Model {
	ti := textinput.New()
	ti.Placeholder = "Commit message..."
	ti.CharLimit = 500
	ti.Width = width - 10

	return Model{
		Repo:        repo,
		Width:       width,
		Height:      height,
		Keys:        keys,
		Config:      cfg,
		ActiveView:  ViewSource,
		CommitInput: ti,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.CommitInput.Width = msg.Width - 10

	case t.StashListMsg:
		if msg.Err == nil {
			m.Stashes = msg.Stashes
		}

	case t.FileTreeMsg:
		if msg.Err == nil {
			m.FileTree = msg.Root
			m.FlatTree = flattenTree(msg.Root)
		}

	case t.LogMsg:
		if msg.Err == nil {
			m.LogEntries = msg.Entries
		}

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	// If typing a commit message, delegate to text input
	if m.Focus == FocusCommit && m.CommitInput.Focused() {
		return m.updateCommitInput(msg)
	}

	switch m.ActiveView {
	case ViewSource:
		return m.handleSourceKey(msg)
	case ViewFiles:
		return m.handleFilesKey(msg)
	case ViewHistory:
		return m.handleHistoryKey(msg)
	}

	return m, nil
}

// ── Source Control Keys ─────────────────────────────────

func (m Model) handleSourceKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.Keys.Down):
		m.moveFileDown()
	case key.Matches(msg, m.Keys.Up):
		m.moveFileUp()
	case key.Matches(msg, m.Keys.Tab):
		m.cycleFocus()
	case key.Matches(msg, m.Keys.Space):
		return m.toggleStage()
	case key.Matches(msg, m.Keys.StageAll):
		return m.stageAll()
	case key.Matches(msg, m.Keys.UnstageAll):
		return m.unstageAll()
	case key.Matches(msg, m.Keys.Commit):
		m.CommitInput.Focus()
		m.Focus = FocusCommit
	case key.Matches(msg, m.Keys.Push):
		return m.push()
	case key.Matches(msg, m.Keys.Pull):
		return m.pull()
	case key.Matches(msg, m.Keys.FetchKey):
		return m.fetch()
	case key.Matches(msg, m.Keys.Discard):
		return m.discardFile()
	case key.Matches(msg, m.Keys.DiscardAll):
		return m.discardAll()
	case key.Matches(msg, m.Keys.OpenEditor):
		return m.openEditor()
	case key.Matches(msg, m.Keys.JumpTop):
		m.FocusedFile = 0
	case key.Matches(msg, m.Keys.JumpBottom):
		total := m.totalFiles()
		if total > 0 {
			m.FocusedFile = total - 1
		}
	case key.Matches(msg, m.Keys.HalfPageDown):
		for i := 0; i < m.Height/2; i++ {
			m.moveFileDown()
		}
	case key.Matches(msg, m.Keys.HalfPageUp):
		for i := 0; i < m.Height/2; i++ {
			m.moveFileUp()
		}
	}
	return m, nil
}

func (m Model) updateCommitInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.CommitInput.Blur()
		m.Focus = FocusStaged
		return m, nil
	case "alt+s":
		return m.submitCommit(false, false)
	case "alt+a":
		return m.submitCommit(false, true)
	case "alt+p":
		return m.submitCommit(true, false)
	default:
		var cmd tea.Cmd
		m.CommitInput, cmd = m.CommitInput.Update(msg)
		return m, cmd
	}
}

func (m Model) submitCommit(andPush, amend bool) (Model, tea.Cmd) {
	message := m.CommitInput.Value()
	if message == "" {
		return m, nil
	}
	repoPath := m.Repo.Path

	m.CommitInput.SetValue("")
	m.CommitInput.Blur()
	m.Focus = FocusStaged

	return m, func() tea.Msg {
		var err error
		if amend {
			err = git.CommitAmend(repoPath, message)
		} else {
			err = git.Commit(repoPath, message)
		}
		if err != nil {
			return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "commit", Err: err}
		}
		if andPush {
			output, pushErr := git.Push(repoPath)
			return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "commit+push", Err: pushErr, Output: output}
		}
		return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "commit", Err: nil}
	}
}

func (m *Model) moveFileDown() {
	total := m.totalFiles()
	if total == 0 {
		return
	}
	if m.FocusedFile < total-1 {
		m.FocusedFile++
	}
}

func (m *Model) moveFileUp() {
	if m.FocusedFile > 0 {
		m.FocusedFile--
	}
}

func (m *Model) cycleFocus() {
	sections := []FocusSection{FocusCommit, FocusStaged, FocusUnstaged, FocusStashes}
	for i, s := range sections {
		if s == m.Focus {
			m.Focus = sections[(i+1)%len(sections)]
			m.FocusedFile = 0
			if m.Focus == FocusCommit {
				m.CommitInput.Focus()
			} else {
				m.CommitInput.Blur()
			}
			return
		}
	}
}

func (m Model) totalFiles() int {
	if m.Repo == nil {
		return 0
	}
	return len(m.Repo.Staged) + len(m.Repo.Unstaged) + len(m.Repo.Untracked)
}

func (m Model) focusedFileChange() (*git.FileChange, string) {
	if m.Repo == nil {
		return nil, ""
	}
	idx := m.FocusedFile

	if m.Focus == FocusStaged || (m.Focus != FocusUnstaged && idx < len(m.Repo.Staged)) {
		if idx < len(m.Repo.Staged) {
			return &m.Repo.Staged[idx], "staged"
		}
	}

	// Map global index
	if idx < len(m.Repo.Staged) {
		return &m.Repo.Staged[idx], "staged"
	}
	idx -= len(m.Repo.Staged)
	if idx < len(m.Repo.Unstaged) {
		return &m.Repo.Unstaged[idx], "unstaged"
	}
	idx -= len(m.Repo.Unstaged)
	if idx < len(m.Repo.Untracked) {
		return &m.Repo.Untracked[idx], "untracked"
	}
	return nil, ""
}

func (m Model) toggleStage() (Model, tea.Cmd) {
	fc, section := m.focusedFileChange()
	if fc == nil {
		return m, nil
	}
	repoPath := m.Repo.Path
	filePath := fc.Path

	switch section {
	case "staged":
		return m, func() tea.Msg {
			err := git.UnstageFile(repoPath, filePath)
			return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "unstage", Err: err}
		}
	case "unstaged", "untracked":
		return m, func() tea.Msg {
			err := git.StageFile(repoPath, filePath)
			return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "stage", Err: err}
		}
	}
	return m, nil
}

func (m Model) stageAll() (Model, tea.Cmd) {
	repoPath := m.Repo.Path
	return m, func() tea.Msg {
		err := git.StageAll(repoPath)
		return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "stage", Err: err}
	}
}

func (m Model) unstageAll() (Model, tea.Cmd) {
	repoPath := m.Repo.Path
	return m, func() tea.Msg {
		err := git.UnstageAll(repoPath)
		return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "unstage", Err: err}
	}
}

func (m Model) push() (Model, tea.Cmd) {
	repoPath := m.Repo.Path
	return m, func() tea.Msg {
		output, err := git.Push(repoPath)
		return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "push", Err: err, Output: output}
	}
}

func (m Model) pull() (Model, tea.Cmd) {
	repoPath := m.Repo.Path
	return m, func() tea.Msg {
		output, err := git.Pull(repoPath)
		return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "pull", Err: err, Output: output}
	}
}

func (m Model) fetch() (Model, tea.Cmd) {
	repoPath := m.Repo.Path
	return m, func() tea.Msg {
		output, err := git.Fetch(repoPath)
		return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "fetch", Err: err, Output: output}
	}
}

func (m Model) discardFile() (Model, tea.Cmd) {
	fc, section := m.focusedFileChange()
	if fc == nil {
		return m, nil
	}
	repoPath := m.Repo.Path
	filePath := fc.Path
	isUntracked := section == "untracked"

	return m, func() tea.Msg {
		err := git.DiscardFile(repoPath, filePath, isUntracked)
		return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "discard", Err: err}
	}
}

func (m Model) discardAll() (Model, tea.Cmd) {
	repoPath := m.Repo.Path
	return m, func() tea.Msg {
		err := git.DiscardAll(repoPath)
		return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "discard", Err: err}
	}
}

func (m Model) openEditor() (Model, tea.Cmd) {
	fc, _ := m.focusedFileChange()
	if fc == nil {
		return m, nil
	}
	filePath := m.Repo.Path + "/" + fc.Path
	return m, editor.Open(filePath, m.Config)
}

// ── File Explorer Keys ──────────────────────────────────

func (m Model) handleFilesKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.Keys.Down):
		if m.TreeCursor < len(m.FlatTree)-1 {
			m.TreeCursor++
		}
	case key.Matches(msg, m.Keys.Up):
		if m.TreeCursor > 0 {
			m.TreeCursor--
		}
	case key.Matches(msg, m.Keys.Enter):
		return m.toggleOrOpenFile()
	case key.Matches(msg, m.Keys.ExpandDir):
		return m.expandDir()
	case key.Matches(msg, m.Keys.CollapseDir):
		return m.collapseDir()
	case key.Matches(msg, m.Keys.Space):
		return m.stageUnstageTreeFile()
	case key.Matches(msg, m.Keys.JumpTop):
		m.TreeCursor = 0
	case key.Matches(msg, m.Keys.JumpBottom):
		if len(m.FlatTree) > 0 {
			m.TreeCursor = len(m.FlatTree) - 1
		}
	case key.Matches(msg, m.Keys.HalfPageDown):
		m.TreeCursor += m.Height / 2
		if len(m.FlatTree) > 0 && m.TreeCursor >= len(m.FlatTree) {
			m.TreeCursor = len(m.FlatTree) - 1
		}
	case key.Matches(msg, m.Keys.HalfPageUp):
		m.TreeCursor -= m.Height / 2
		if m.TreeCursor < 0 {
			m.TreeCursor = 0
		}
	}
	return m, nil
}

func (m Model) toggleOrOpenFile() (Model, tea.Cmd) {
	if m.TreeCursor >= len(m.FlatTree) {
		return m, nil
	}
	node := m.FlatTree[m.TreeCursor]
	if node.IsDir {
		node.IsExpanded = !node.IsExpanded
		m.FlatTree = flattenTree(m.FileTree)
		return m, nil
	}
	// Open file in editor
	filePath := m.Repo.Path + "/" + node.Path
	return m, editor.Open(filePath, m.Config)
}

func (m Model) expandDir() (Model, tea.Cmd) {
	if m.TreeCursor >= len(m.FlatTree) {
		return m, nil
	}
	node := m.FlatTree[m.TreeCursor]
	if node.IsDir && !node.IsExpanded {
		node.IsExpanded = true
		m.FlatTree = flattenTree(m.FileTree)
	}
	return m, nil
}

func (m Model) collapseDir() (Model, tea.Cmd) {
	if m.TreeCursor >= len(m.FlatTree) {
		return m, nil
	}
	node := m.FlatTree[m.TreeCursor]
	if node.IsDir && node.IsExpanded {
		node.IsExpanded = false
		m.FlatTree = flattenTree(m.FileTree)
	}
	return m, nil
}

func (m Model) stageUnstageTreeFile() (Model, tea.Cmd) {
	if m.TreeCursor >= len(m.FlatTree) {
		return m, nil
	}
	node := m.FlatTree[m.TreeCursor]
	if node.IsDir || node.Status == nil {
		return m, nil
	}

	repoPath := m.Repo.Path
	filePath := node.Path

	if node.IsStaged {
		return m, func() tea.Msg {
			err := git.UnstageFile(repoPath, filePath)
			return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "unstage", Err: err}
		}
	}
	return m, func() tea.Msg {
		err := git.StageFile(repoPath, filePath)
		return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "stage", Err: err}
	}
}

// ── History Keys ────────────────────────────────────────

func (m Model) handleHistoryKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.Keys.Down):
		if m.LogCursor < len(m.LogEntries)-1 {
			m.LogCursor++
		}
	case key.Matches(msg, m.Keys.Up):
		if m.LogCursor > 0 {
			m.LogCursor--
		}
	case key.Matches(msg, m.Keys.CherryPick):
		return m.cherryPick()
	case key.Matches(msg, m.Keys.Revert):
		return m.revertCommit()
	case key.Matches(msg, m.Keys.JumpTop):
		m.LogCursor = 0
	case key.Matches(msg, m.Keys.JumpBottom):
		if len(m.LogEntries) > 0 {
			m.LogCursor = len(m.LogEntries) - 1
		}
	case key.Matches(msg, m.Keys.HalfPageDown):
		m.LogCursor += m.Height / 2
		if len(m.LogEntries) > 0 && m.LogCursor >= len(m.LogEntries) {
			m.LogCursor = len(m.LogEntries) - 1
		}
	case key.Matches(msg, m.Keys.HalfPageUp):
		m.LogCursor -= m.Height / 2
		if m.LogCursor < 0 {
			m.LogCursor = 0
		}
	}
	return m, nil
}

func (m Model) cherryPick() (Model, tea.Cmd) {
	if m.LogCursor >= len(m.LogEntries) {
		return m, nil
	}
	entry := m.LogEntries[m.LogCursor]
	repoPath := m.Repo.Path
	return m, func() tea.Msg {
		output, err := git.CherryPick(repoPath, entry.FullHash)
		return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "cherry-pick", Err: err, Output: output}
	}
}

func (m Model) revertCommit() (Model, tea.Cmd) {
	if m.LogCursor >= len(m.LogEntries) {
		return m, nil
	}
	entry := m.LogEntries[m.LogCursor]
	repoPath := m.Repo.Path
	return m, func() tea.Msg {
		output, err := git.RevertCommit(repoPath, entry.FullHash)
		return t.GitOpCompleteMsg{RepoPath: repoPath, Op: "revert", Err: err, Output: output}
	}
}

// ── View ────────────────────────────────────────────────

// CursorLine returns the line number (0-based) of the focused item in the rendered output.
func (m Model) CursorLine() int {
	switch m.ActiveView {
	case ViewFiles:
		return 1 + m.TreeCursor // 1 for leading \n
	case ViewHistory:
		return 1 + m.LogCursor // 1 for leading \n
	default:
		return m.cursorLineSource()
	}
}

func (m Model) cursorLineSource() int {
	// Commit input section: \n, COMMIT MESSAGE, input, alt+s line = 4 lines
	line := 4

	if len(m.Repo.Staged) > 0 {
		line++ // \n
		line++ // STAGED CHANGES (N)
		for i := range m.Repo.Staged {
			if m.Focus == FocusStaged && m.FocusedFile == i {
				return line
			}
			line++
		}
	}

	if len(m.Repo.Unstaged) > 0 || len(m.Repo.Untracked) > 0 {
		line++ // \n
		line++ // CHANGES (N)
		for i := range m.Repo.Unstaged {
			if m.Focus == FocusUnstaged && m.FocusedFile == i {
				return line
			}
			line++
		}
		for i := range m.Repo.Untracked {
			if m.Focus == FocusUnstaged && m.FocusedFile == len(m.Repo.Unstaged)+i {
				return line
			}
			line++
		}
	}

	if len(m.Stashes) > 0 {
		line++ // \n
		line++ // STASHES (N)
		for i := range m.Stashes {
			if m.Focus == FocusStashes && m.FocusedFile == i {
				return line
			}
			line++
		}
	}

	return line
}

// View renders the drill-in mode.
func (m Model) View() string {
	switch m.ActiveView {
	case ViewFiles:
		return m.viewFiles()
	case ViewHistory:
		return m.viewHistory()
	default:
		return m.viewSource()
	}
}

func (m Model) viewSource() string {
	var b strings.Builder

	// Commit input
	b.WriteString("\n")
	b.WriteString("  " + styles.CommitLabelStyle.Render("COMMIT MESSAGE") + "\n")
	inputWidth := m.Width - 6
	if inputWidth < 20 {
		inputWidth = 20
	}
	b.WriteString("  " + styles.CommitInputStyle.MaxWidth(inputWidth).Render(m.CommitInput.View()) + "\n")
	b.WriteString("  " + styles.FooterStyle.Render("alt+s commit staged  alt+a amend  alt+p commit + push") + "\n")

	// Staged changes
	if len(m.Repo.Staged) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.StagedHeaderStyle.Render(fmt.Sprintf("STAGED CHANGES (%d)", len(m.Repo.Staged))) + "\n")
		for i, f := range m.Repo.Staged {
			focused := m.Focus == FocusStaged && m.FocusedFile == i
			b.WriteString(m.renderSourceFile(f, focused, true))
		}
	}

	// Unstaged changes
	if len(m.Repo.Unstaged) > 0 || len(m.Repo.Untracked) > 0 {
		b.WriteString("\n")
		total := len(m.Repo.Unstaged) + len(m.Repo.Untracked)
		b.WriteString(styles.UnstagedHeaderStyle.Render(fmt.Sprintf("CHANGES (%d)", total)) + "\n")
		for i, f := range m.Repo.Unstaged {
			focused := m.Focus == FocusUnstaged && m.FocusedFile == i
			b.WriteString(m.renderSourceFile(f, focused, false))
		}
		for i, f := range m.Repo.Untracked {
			focused := m.Focus == FocusUnstaged && m.FocusedFile == len(m.Repo.Unstaged)+i
			b.WriteString(m.renderSourceFile(f, focused, false))
		}
	}

	// Stashes
	if len(m.Stashes) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.StashStyle.Render(fmt.Sprintf("STASHES (%d)", len(m.Stashes))) + "\n")
		for i, s := range m.Stashes {
			focused := m.Focus == FocusStashes && m.FocusedFile == i
			cursor := styles.Cursor(focused)
			line := fmt.Sprintf("%sstash@{%d} %s", cursor, s.Index, s.Message)
			timeStr := styles.DateStyle.Render(styles.RelativeTime(s.Date))
			line += "  " + timeStr
			if focused {
				b.WriteString(styles.SelectedStyle.Render(line) + "\n")
			} else {
				b.WriteString(line + "\n")
			}
		}
	}

	if m.StatusMsg != "" {
		b.WriteString("\n  " + m.StatusMsg + "\n")
	}

	return b.String()
}

func (m Model) renderSourceFile(f git.FileChange, focused, staged bool) string {
	cursor := styles.Cursor(focused)
	status := styles.StatusStyle(f.Status).Render(styles.StatusChar(f.Status))

	// Truncate path if needed to leave room for right-side info
	maxPathWidth := m.Width - 22 // reserve space for cursor + status + margins
	if maxPathWidth < 10 {
		maxPathWidth = 10
	}
	path := styles.Truncate(f.Path, maxPathWidth)

	line := fmt.Sprintf("%s%s  %s", cursor, status, path)

	// Only show right-side details if we have enough room
	if m.Width > 60 {
		var right string
		if f.Insertions > 0 || f.Deletions > 0 {
			right += "  " + styles.AddedStyle.Render(fmt.Sprintf("+%d", f.Insertions)) + " " +
				styles.DeletedStyle.Render(fmt.Sprintf("-%d", f.Deletions))
		}
		if staged {
			right += "  " + styles.FooterStyle.Render("[u]nstage")
		} else {
			right += "  " + styles.FooterStyle.Render("[s]tage [d]iscard")
		}

		leftW := lipgloss.Width(line)
		rightW := lipgloss.Width(right)
		gap := m.Width - leftW - rightW
		if gap < 1 {
			gap = 1
		}
		line += strings.Repeat(" ", gap) + right
	}

	if focused {
		return styles.SelectedStyle.Render(line) + "\n"
	}
	return line + "\n"
}

func (m Model) viewFiles() string {
	if m.FileTree == nil {
		return "\n  Loading file tree...\n"
	}

	var b strings.Builder
	b.WriteString("\n")

	for i, node := range m.FlatTree {
		focused := i == m.TreeCursor
		b.WriteString(m.renderTreeNode(node, focused))
	}

	return b.String()
}

func (m Model) renderTreeNode(node *git.FileTreeNode, focused bool) string {
	cursor := styles.Cursor(focused)
	indent := strings.Repeat("  ", node.Depth)

	var icon string
	if node.IsDir {
		if node.IsExpanded {
			icon = styles.ExpandedIcon + " "
		} else {
			icon = styles.CollapsedIcon + " "
		}
	} else {
		icon = "  "
	}

	name := node.Name
	if node.IsDir {
		name += "/"
	}

	var statusStr string
	if node.Status != nil {
		s := *node.Status
		statusChar := styles.StatusChar(s)
		var label string
		switch {
		case node.IsStaged:
			label = "staged"
		case s == git.StatusUntracked:
			label = "new"
		default:
			label = "modified"
		}
		statusStr = styles.StatusStyle(s).Render(statusChar + " " + label)
	}

	if node.IsIgnored {
		name = styles.UntrackedStyle.Render(name)
		statusStr = styles.UntrackedStyle.Render("ignored")
	}

	line := cursor + indent + icon + name
	if statusStr != "" {
		leftW := lipgloss.Width(line)
		rightW := lipgloss.Width(statusStr)
		gap := m.Width - leftW - rightW - 2
		if gap < 1 {
			gap = 1
		}
		line += strings.Repeat(" ", gap) + statusStr
	}

	if focused {
		return styles.SelectedStyle.Render(line) + "\n"
	}
	return line + "\n"
}

func (m Model) viewHistory() string {
	if len(m.LogEntries) == 0 {
		return "\n  No commits yet.\n"
	}

	var b strings.Builder
	b.WriteString("\n")

	for i, entry := range m.LogEntries {
		focused := i == m.LogCursor
		cursor := styles.Cursor(focused)

		hash := styles.HashStyle.Render(entry.Hash)
		msg := entry.Message
		timeStr := styles.DateStyle.Render(styles.RelativeTime(entry.Date))
		author := styles.AuthorStyle.Render(entry.Author)

		var refs string
		for _, ref := range entry.Refs {
			refs += " " + styles.RefStyle.Render(ref)
		}

		// Adapt message width to terminal size
		msgWidth := m.Width/2 - 15
		if msgWidth < 15 {
			msgWidth = 15
		}
		line := fmt.Sprintf("%s%s  %s%s", cursor, hash, styles.Truncate(msg, msgWidth), refs)

		var fullLine string
		if m.Width > 70 {
			right := fmt.Sprintf("%s  %s", timeStr, author)
			leftW := lipgloss.Width(line)
			rightW := lipgloss.Width(right)
			gap := m.Width - leftW - rightW
			if gap < 1 {
				gap = 1
			}
			fullLine = line + strings.Repeat(" ", gap) + right
		} else {
			fullLine = line
		}

		if focused {
			b.WriteString(styles.SelectedStyle.Render(fullLine) + "\n")
		} else {
			b.WriteString(fullLine + "\n")
		}
	}

	return b.String()
}

// flattenTree converts a tree into a flat list of visible nodes.
func flattenTree(root *git.FileTreeNode) []*git.FileTreeNode {
	if root == nil {
		return nil
	}
	var result []*git.FileTreeNode
	flattenNode(root, &result)
	return result
}

func flattenNode(node *git.FileTreeNode, result *[]*git.FileTreeNode) {
	*result = append(*result, node)
	if node.IsDir && node.IsExpanded {
		for _, child := range node.Children {
			flattenNode(child, result)
		}
	}
}
