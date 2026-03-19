package app

import (
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jack-braga/gitto/internal/config"
	"github.com/jack-braga/gitto/internal/git"
	"github.com/jack-braga/gitto/internal/ui/drillin"
	"github.com/jack-braga/gitto/internal/ui/footer"
	"github.com/jack-braga/gitto/internal/ui/header"
	"github.com/jack-braga/gitto/internal/ui/overlays"
	"github.com/jack-braga/gitto/internal/ui/overview"
	"github.com/jack-braga/gitto/internal/ui/styles"
	"github.com/jack-braga/gitto/internal/ui/tabs"
)

// Mode represents the current application mode.
type Mode int

const (
	ModeOverview Mode = iota
	ModeDrillIn
)

// Model is the root Bubble Tea model for gitto.
type Model struct {
	// Core state
	Mode      Mode
	ParentDir string
	RepoPaths []string
	Repos     map[string]*git.Repository
	Config    *config.Config
	Keys      KeyMap
	Width     int
	Height    int
	Ready     bool

	// Sub-models
	Header   header.Model
	Tabs     tabs.Model
	Footer   footer.Model
	Overview overview.Model
	DrillIn  drillin.Model
	Viewport viewport.Model

	// Overlays
	BranchPicker  overlays.BranchPicker
	StashDialog   overlays.StashDialog
	ConfirmDialog overlays.ConfirmDialog
	HelpOverlay   overlays.HelpOverlay

	// Status
	StatusMsg    string
	StatusExpiry time.Time
	NoPoll       bool
}

// New creates the root application model.
func New(parentDir string, cfg *config.Config, noPoll bool) Model {
	keys := DefaultKeyMap()
	return Model{
		ParentDir:    parentDir,
		Config:       cfg,
		Keys:         keys,
		Repos:        make(map[string]*git.Repository),
		Header:       header.New(parentDir),
		Tabs:         tabs.New(cfg.DefaultView),
		Footer:       footer.New(),
		Overview:     overview.New(keys),
		BranchPicker: overlays.NewBranchPicker(),
		StashDialog:  overlays.NewStashDialog(),
		NoPoll:       noPoll,
	}
}

// Init starts repo discovery and the background tick.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.discoverRepos(),
		m.tick(),
	)
}

// Update handles all messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = true

		m.Header.Width = msg.Width
		m.Tabs.Width = msg.Width
		m.Footer.Width = msg.Width
		m.BranchPicker.Width = msg.Width
		m.StashDialog.Width = msg.Width
		m.HelpOverlay.Width = msg.Width

		// Set up viewport with available content height
		contentHeight := styles.ContentHeight(msg.Height)
		m.Viewport = viewport.New(msg.Width, contentHeight)
		m.Viewport.Style = lipglossNoop()
		// Disable viewport's default keybindings — we handle navigation ourselves
		m.Viewport.KeyMap = viewport.KeyMap{}

		// Propagate to sub-models
		m.Overview.Width = msg.Width
		m.Overview.Height = contentHeight
		if m.Mode == ModeDrillIn {
			m.DrillIn.Width = msg.Width
			m.DrillIn.Height = contentHeight
		}

	case tea.KeyMsg:
		return m.handleKey(msg)

	case ReposDiscoveredMsg:
		m.RepoPaths = msg.Repos
		cmds = append(cmds, m.refreshAllRepos())

	case BatchStatusMsg:
		for _, s := range msg.Statuses {
			if s.Err == nil && s.Status != nil {
				m.Repos[s.Path] = s.Status
			}
		}
		m.syncOverviewRepos()
		if m.Mode == ModeDrillIn && m.DrillIn.Repo != nil {
			if updated, ok := m.Repos[m.DrillIn.Repo.Path]; ok {
				m.DrillIn.Repo = updated
			}
		}

	case RepoStatusMsg:
		if msg.Err == nil && msg.Status != nil {
			m.Repos[msg.Path] = msg.Status
			m.syncOverviewRepos()
			if m.Mode == ModeDrillIn && m.DrillIn.Repo != nil && m.DrillIn.Repo.Path == msg.Path {
				m.DrillIn.Repo = msg.Status
			}
		}

	case GitOpCompleteMsg:
		if msg.Err != nil {
			m.setStatus("Error: " + msg.Err.Error())
		} else {
			m.setStatus(msg.Op + " completed")
		}
		cmds = append(cmds, m.refreshRepo(msg.RepoPath))

	case BranchListMsg:
		if msg.Err == nil {
			m.BranchPicker.Show(msg.RepoPath, msg.Branches)
		}

	case StashListMsg:
		// User explicitly opened stash dialog (pressed z)
		if msg.Err == nil {
			m.StashDialog.Show(msg.RepoPath, msg.Stashes)
			if m.Mode == ModeDrillIn {
				m.DrillIn.Stashes = msg.Stashes
			}
		}

	case StashDataMsg:
		// Background stash data load (drill-in entry) — no dialog
		if msg.Err == nil && m.Mode == ModeDrillIn {
			m.DrillIn.Stashes = msg.Stashes
		}

	case LogMsg:
		if m.Mode == ModeDrillIn {
			m.DrillIn, _ = m.DrillIn.Update(msg)
		}

	case FileTreeMsg:
		if m.Mode == ModeDrillIn {
			m.DrillIn, _ = m.DrillIn.Update(msg)
		}

	case TickMsg:
		if !m.NoPoll {
			cmds = append(cmds, m.refreshAllRepos())
			cmds = append(cmds, m.tick())
		}

	case overlays.ConfirmResult:
		// Handle confirmed actions

	case DrillInMsg:
		m.enterDrillIn(msg.RepoIndex)
		cmds = append(cmds, m.loadDrillInData())
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Overlays consume keys first
	if m.HelpOverlay.Active {
		var cmd tea.Cmd
		m.HelpOverlay, cmd = m.HelpOverlay.Update(msg)
		return m, cmd
	}
	if m.ConfirmDialog.Active {
		var cmd tea.Cmd
		m.ConfirmDialog, cmd = m.ConfirmDialog.Update(msg)
		return m, cmd
	}
	if m.BranchPicker.Active {
		var cmd tea.Cmd
		m.BranchPicker, cmd = m.BranchPicker.Update(msg)
		return m, cmd
	}
	if m.StashDialog.Active {
		var cmd tea.Cmd
		m.StashDialog, cmd = m.StashDialog.Update(msg)
		return m, cmd
	}

	// If a text input is focused, skip global keys — let the sub-model handle them
	if m.isTextInputActive() {
		var cmd tea.Cmd
		switch m.Mode {
		case ModeOverview:
			m.Overview, cmd = m.Overview.Update(msg)
		case ModeDrillIn:
			m.DrillIn, cmd = m.DrillIn.Update(msg)
		}
		return m, cmd
	}

	// Global keys
	switch {
	case key.Matches(msg, m.Keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.Keys.Help):
		m.HelpOverlay.Toggle()
		return m, nil
	case key.Matches(msg, m.Keys.ForceRefresh):
		return m, m.refreshAllRepos()
	case key.Matches(msg, m.Keys.SourceView):
		m.Tabs.ActiveTab = tabs.Source
		m.Footer.ActiveView = tabs.Source
		if m.Mode == ModeDrillIn {
			m.DrillIn.ActiveView = drillin.ViewSource
		}
		return m, nil
	case key.Matches(msg, m.Keys.FilesView):
		m.Tabs.ActiveTab = tabs.Files
		m.Footer.ActiveView = tabs.Files
		if m.Mode == ModeDrillIn {
			m.DrillIn.ActiveView = drillin.ViewFiles
			if m.DrillIn.FileTree == nil {
				return m, m.loadFileTree()
			}
		}
		return m, nil
	case key.Matches(msg, m.Keys.HistoryView):
		// History only available in drill-in mode
		if m.Mode == ModeDrillIn {
			m.Tabs.ActiveTab = tabs.History
			m.Footer.ActiveView = tabs.History
			m.DrillIn.ActiveView = drillin.ViewHistory
			if len(m.DrillIn.LogEntries) == 0 {
				return m, m.loadLog()
			}
		}
		return m, nil
	case key.Matches(msg, m.Keys.Escape):
		if m.Mode == ModeDrillIn {
			m.exitDrillIn()
			return m, nil
		}
	case key.Matches(msg, m.Keys.BranchKey):
		return m, m.openBranchPicker()
	case key.Matches(msg, m.Keys.StashKey):
		return m, m.openStashDialog()
	}

	// Delegate to mode-specific model
	var cmd tea.Cmd
	switch m.Mode {
	case ModeOverview:
		m.Overview, cmd = m.Overview.Update(msg)
	case ModeDrillIn:
		m.DrillIn, cmd = m.DrillIn.Update(msg)
	}

	return m, cmd
}

// View renders the entire application.
func (m Model) View() string {
	if !m.Ready {
		return "\n  Initializing...\n"
	}

	// Min-size guard
	if m.Width < styles.MinWidth || m.Height < styles.MinHeight {
		return styles.TooSmall(m.Width, m.Height)
	}

	var b strings.Builder

	// Header
	b.WriteString(styles.ClampLine(m.Header.View(), m.Width) + "\n")
	b.WriteString(styles.Divider(m.Width) + "\n")

	// Tabs
	b.WriteString(styles.ClampLine(m.Tabs.View(), m.Width) + "\n")
	b.WriteString(styles.Divider(m.Width) + "\n")

	// Content area — rendered into viewport for scrolling
	var content string

	if m.HelpOverlay.Active {
		content = m.HelpOverlay.View()
	} else if m.BranchPicker.Active {
		content = m.BranchPicker.View()
	} else if m.StashDialog.Active {
		content = m.StashDialog.View()
	} else if m.ConfirmDialog.Active {
		content = m.ConfirmDialog.View()
	} else {
		switch m.Mode {
		case ModeOverview:
			content = m.Overview.View()
		case ModeDrillIn:
			content = m.DrillIn.View()
		}

		// Status message
		if m.StatusMsg != "" && time.Now().Before(m.StatusExpiry) {
			content += "\n  " + styles.FooterStyle.Render(m.StatusMsg) + "\n"
		}
	}

	// Clamp each line to terminal width
	content = clampContent(content, m.Width)

	// Fit content into viewport
	contentHeight := styles.ContentHeight(m.Height)
	m.Viewport.Width = m.Width
	m.Viewport.Height = contentHeight
	m.Viewport.SetContent(content)
	b.WriteString(m.Viewport.View() + "\n")

	// Footer
	b.WriteString(styles.Divider(m.Width) + "\n")
	b.WriteString(styles.ClampLine(m.Footer.View(), m.Width))

	return b.String()
}

// clampContent truncates each line in a multi-line string to maxWidth.
func clampContent(content string, maxWidth int) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = styles.ClampLine(line, maxWidth)
	}
	return strings.Join(lines, "\n")
}

// lipglossNoop returns an empty lipgloss style.
func lipglossNoop() lipgloss.Style {
	return lipgloss.NewStyle()
}

// isTextInputActive returns true if a text input is currently focused,
// meaning global keybindings should be bypassed so the user can type.
func (m Model) isTextInputActive() bool {
	if m.Mode == ModeOverview && m.Overview.Focus == overview.FocusCommitInput {
		return true
	}
	if m.Mode == ModeDrillIn && m.DrillIn.CommitInput.Focused() {
		return true
	}
	return false
}

// ── Helper methods ──────────────────────────────────────

func (m Model) discoverRepos() tea.Cmd {
	parentDir := m.ParentDir
	return func() tea.Msg {
		repos, err := git.DiscoverRepos(parentDir)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return ReposDiscoveredMsg{Repos: repos}
	}
}

func (m Model) refreshAllRepos() tea.Cmd {
	paths := m.RepoPaths
	return func() tea.Msg {
		var wg sync.WaitGroup
		results := make([]RepoStatusMsg, len(paths))

		for i, path := range paths {
			wg.Add(1)
			go func(idx int, p string) {
				defer wg.Done()
				status, err := git.GetStatus(p)
				results[idx] = RepoStatusMsg{Path: p, Status: status, Err: err}
			}(i, path)
		}

		wg.Wait()
		return BatchStatusMsg{Statuses: results}
	}
}

func (m Model) refreshRepo(repoPath string) tea.Cmd {
	return func() tea.Msg {
		status, err := git.GetStatus(repoPath)
		return RepoStatusMsg{Path: repoPath, Status: status, Err: err}
	}
}

func (m Model) tick() tea.Cmd {
	if m.NoPoll {
		return nil
	}
	interval := time.Duration(m.Config.PollInterval) * time.Second
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return TickMsg{}
	})
}

func (m *Model) syncOverviewRepos() {
	repos := make([]*git.Repository, 0, len(m.RepoPaths))
	for _, path := range m.RepoPaths {
		if r, ok := m.Repos[path]; ok {
			repos = append(repos, r)
		}
	}
	m.Overview.SetRepos(repos)
}

func (m *Model) enterDrillIn(repoIndex int) {
	repos := m.Overview.Repos
	if repoIndex >= len(repos) {
		return
	}
	repo := repos[repoIndex]

	m.Mode = ModeDrillIn
	m.DrillIn = drillin.New(repo, m.Keys, m.Config, m.Width, styles.ContentHeight(m.Height))
	m.DrillIn.ActiveView = m.Tabs.ActiveTab

	// Update header and tabs
	m.Header.IsDrillIn = true
	m.Tabs.IsDrillIn = true
	m.Header.RepoName = repo.Name
	m.Header.Branch = repo.Branch
	m.Header.Upstream = repo.Upstream
	m.Header.Ahead = repo.Ahead
	m.Header.Behind = repo.Behind

	// Update footer
	m.Footer.Mode = "drillin"
}

func (m *Model) exitDrillIn() {
	m.Mode = ModeOverview
	m.Header.IsDrillIn = false
	m.Tabs.IsDrillIn = false
	m.Footer.Mode = "overview"
	// If we were on history tab, switch back to source (not available in overview)
	if m.Tabs.ActiveTab == tabs.History {
		m.Tabs.ActiveTab = tabs.Source
		m.Footer.ActiveView = tabs.Source
	}
}

func (m *Model) setStatus(msg string) {
	m.StatusMsg = msg
	m.StatusExpiry = time.Now().Add(5 * time.Second)
	if m.Mode == ModeOverview {
		m.Overview.StatusMsg = msg
	} else {
		m.DrillIn.StatusMsg = msg
	}
}

func (m Model) loadDrillInData() tea.Cmd {
	return tea.Batch(
		m.loadStashes(),
	)
}

func (m Model) loadStashes() tea.Cmd {
	if m.DrillIn.Repo == nil {
		return nil
	}
	repoPath := m.DrillIn.Repo.Path
	return func() tea.Msg {
		stashes, err := git.ListStashes(repoPath)
		return StashDataMsg{RepoPath: repoPath, Stashes: stashes, Err: err}
	}
}

func (m Model) loadFileTree() tea.Cmd {
	if m.DrillIn.Repo == nil {
		return nil
	}
	repoPath := m.DrillIn.Repo.Path
	return func() tea.Msg {
		root, err := git.BuildFileTree(repoPath)
		return FileTreeMsg{RepoPath: repoPath, Root: root, Err: err}
	}
}

func (m Model) loadLog() tea.Cmd {
	if m.DrillIn.Repo == nil {
		return nil
	}
	repoPath := m.DrillIn.Repo.Path
	return func() tea.Msg {
		entries, err := git.GetLog(repoPath, 50)
		return LogMsg{RepoPath: repoPath, Entries: entries, Err: err}
	}
}

func (m Model) openBranchPicker() tea.Cmd {
	var repoPath string
	if m.Mode == ModeDrillIn && m.DrillIn.Repo != nil {
		repoPath = m.DrillIn.Repo.Path
	} else {
		repo := m.Overview.FocusedRepository()
		if repo == nil {
			return nil
		}
		repoPath = repo.Path
	}

	return func() tea.Msg {
		branches, err := git.ListBranches(repoPath)
		return BranchListMsg{RepoPath: repoPath, Branches: branches, Err: err}
	}
}

func (m Model) openStashDialog() tea.Cmd {
	var repoPath string
	if m.Mode == ModeDrillIn && m.DrillIn.Repo != nil {
		repoPath = m.DrillIn.Repo.Path
	} else {
		repo := m.Overview.FocusedRepository()
		if repo == nil {
			return nil
		}
		repoPath = repo.Path
	}

	return func() tea.Msg {
		stashes, err := git.ListStashes(repoPath)
		return StashListMsg{RepoPath: repoPath, Stashes: stashes, Err: err}
	}
}
