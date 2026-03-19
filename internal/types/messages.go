package types

import (
	"github.com/jack-braga/gitto/internal/git"
)

// ReposDiscoveredMsg is sent when repo discovery completes.
type ReposDiscoveredMsg struct {
	Repos []string
}

// RepoStatusMsg is sent when a single repo's status has been refreshed.
type RepoStatusMsg struct {
	Path   string
	Status *git.Repository
	Err    error
}

// BatchStatusMsg is sent when all repos have been refreshed concurrently.
type BatchStatusMsg struct {
	Statuses []RepoStatusMsg
}

// GitOpCompleteMsg is sent when a git operation completes.
type GitOpCompleteMsg struct {
	RepoPath string
	Op       string
	Err      error
	Output   string
}

// BranchListMsg is sent when branch list is fetched.
type BranchListMsg struct {
	RepoPath string
	Branches []git.Branch
	Err      error
}

// StashListMsg is sent when stash list is fetched.
type StashListMsg struct {
	RepoPath string
	Stashes  []git.Stash
	Err      error
}

// LogMsg is sent when commit log is fetched.
type LogMsg struct {
	RepoPath string
	Entries  []git.LogEntry
	Err      error
}

// FileTreeMsg is sent when file tree is built.
type FileTreeMsg struct {
	RepoPath string
	Root     *git.FileTreeNode
	Err      error
}

// TickMsg is sent on periodic tick for background status refresh.
type TickMsg struct{}

// StatusNotificationMsg is sent when a repo changes unexpectedly.
type StatusNotificationMsg struct {
	RepoPath string
	RepoName string
}

// ErrMsg wraps an error for display.
type ErrMsg struct {
	Err error
}

// DrillInMsg is sent when the user wants to drill into a repo.
type DrillInMsg struct {
	RepoIndex int
}
