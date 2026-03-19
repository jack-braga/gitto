package git

import "time"

// FileStatus represents the git status of a file.
type FileStatus int

const (
	StatusModified  FileStatus = iota // M
	StatusAdded                       // A
	StatusDeleted                     // D
	StatusRenamed                     // R
	StatusCopied                      // C
	StatusUntracked                   // ?
	StatusConflicted                  // U
)

// Repository represents the state of a single git repository.
type Repository struct {
	Name       string
	Path       string
	Branch     string
	Upstream   string
	Ahead      int
	Behind     int
	Staged     []FileChange
	Unstaged   []FileChange
	Untracked  []FileChange
	Stashes    []Stash
	IsLoading  bool
	LastUpdate time.Time
}

// TotalChanges returns the count of all non-clean files.
func (r *Repository) TotalChanges() int {
	return len(r.Staged) + len(r.Unstaged) + len(r.Untracked)
}

// IsClean returns true if there are no changes.
func (r *Repository) IsClean() bool {
	return r.TotalChanges() == 0
}

// FileChange represents a single file change in a repository.
type FileChange struct {
	Path       string
	OldPath    string // For renames
	Status     FileStatus
	Insertions int
	Deletions  int
}

// Stash represents a single stash entry.
type Stash struct {
	Index   int
	Message string
	Date    time.Time
}

// Branch represents a git branch.
type Branch struct {
	Name        string
	IsRemote    bool
	IsCurrent   bool
	Upstream    string
	LastCommit  time.Time
	LastMessage string
}

// LogEntry represents a single commit in the log.
type LogEntry struct {
	Hash     string
	FullHash string
	Message  string
	Author   string
	Date     time.Time
	Refs     []string
	RepoName string
}

// FileTreeNode represents a node in the file tree.
type FileTreeNode struct {
	Name       string
	Path       string
	IsDir      bool
	Children   []*FileTreeNode
	Status     *FileStatus
	IsStaged   bool
	IsExpanded bool
	Depth      int
	IsIgnored  bool
}
