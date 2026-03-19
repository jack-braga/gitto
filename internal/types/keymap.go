package types

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the application.
type KeyMap struct {
	// Global
	Quit         key.Binding
	Help         key.Binding
	ForceRefresh key.Binding
	SourceView   key.Binding
	FilesView    key.Binding
	HistoryView  key.Binding
	Up           key.Binding
	Down         key.Binding

	// Navigation
	Enter    key.Binding
	Escape   key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	DrillIn  key.Binding

	// Source control
	Space         key.Binding
	StageAll      key.Binding
	UnstageAll    key.Binding
	Commit        key.Binding
	CommitSubmit  key.Binding
	AmendCommit   key.Binding
	CommitAndPush key.Binding
	Push          key.Binding
	Pull          key.Binding
	FetchKey      key.Binding
	BranchKey     key.Binding
	StashKey      key.Binding
	MergeKey      key.Binding
	RebaseKey     key.Binding
	OpenEditor    key.Binding
	Discard       key.Binding
	DiscardAll    key.Binding

	// File explorer
	ExpandDir   key.Binding
	CollapseDir key.Binding
	CopyPath    key.Binding
	JumpTop     key.Binding
	JumpBottom  key.Binding

	// History
	CherryPick key.Binding
	Revert     key.Binding
	CopyHash   key.Binding

	// Overlays
	Filter       key.Binding
	NewBranch    key.Binding
	DeleteBranch key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:         key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		Help:         key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		ForceRefresh: key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "refresh all")),
		SourceView:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "source control")),
		FilesView:    key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "files")),
		HistoryView:  key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "history")),
		Up:           key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
		Down:         key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),

		Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "expand/collapse")),
		Escape:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next section")),
		ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev section")),
		DrillIn:  key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open repo")),

		Space:         key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "stage/unstage")),
		StageAll:      key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "stage all")),
		UnstageAll:    key.NewBinding(key.WithKeys("U"), key.WithHelp("U", "unstage all")),
		Commit:        key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "commit")),
		CommitSubmit:  key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "submit commit")),
		AmendCommit:   key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "amend")),
		CommitAndPush: key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "commit & push")),
		Push:          key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "push")),
		Pull:          key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "pull")),
		FetchKey:      key.NewBinding(key.WithKeys("F"), key.WithHelp("F", "fetch")),
		BranchKey:     key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "branches")),
		StashKey:      key.NewBinding(key.WithKeys("z"), key.WithHelp("z", "stash")),
		MergeKey:      key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "merge")),
		RebaseKey:     key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rebase")),
		OpenEditor:    key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "open in editor")),
		Discard:       key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "discard")),
		DiscardAll:    key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "discard all")),

		ExpandDir:   key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "expand")),
		CollapseDir: key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "collapse")),
		CopyPath:    key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy path")),
		JumpTop:     key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
		JumpBottom:  key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),

		CherryPick: key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "cherry-pick")),
		Revert:     key.NewBinding(key.WithKeys("V"), key.WithHelp("V", "revert")),
		CopyHash:   key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy hash")),

		Filter:       key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		NewBranch:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new branch")),
		DeleteBranch: key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "delete branch")),
	}
}
