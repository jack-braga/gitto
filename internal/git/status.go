package git

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GetStatus runs git status and returns a Repository with parsed results.
func GetStatus(repoPath string) (*Repository, error) {
	output, err := GitExec(repoPath, "status", "--porcelain=v2", "--branch")
	if err != nil {
		return nil, fmt.Errorf("getting status for %s: %w", repoPath, err)
	}

	repo, err := ParsePorcelainV2(output)
	if err != nil {
		return nil, fmt.Errorf("parsing status for %s: %w", repoPath, err)
	}

	repo.Path = repoPath
	repo.Name = filepath.Base(repoPath)
	repo.LastUpdate = time.Now()

	return repo, nil
}

// ParsePorcelainV2 parses git status --porcelain=v2 --branch output.
func ParsePorcelainV2(output string) (*Repository, error) {
	repo := &Repository{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "# ") {
			parseBranchHeader(repo, line)
			continue
		}

		if strings.HasPrefix(line, "1 ") {
			parseOrdinaryEntry(repo, line)
			continue
		}

		if strings.HasPrefix(line, "2 ") {
			parseRenameEntry(repo, line)
			continue
		}

		if strings.HasPrefix(line, "? ") {
			path := line[2:]
			repo.Untracked = append(repo.Untracked, FileChange{
				Path:   path,
				Status: StatusUntracked,
			})
			continue
		}

		if strings.HasPrefix(line, "u ") {
			parseUnmergedEntry(repo, line)
			continue
		}
	}

	return repo, nil
}

func parseBranchHeader(repo *Repository, line string) {
	switch {
	case strings.HasPrefix(line, "# branch.head "):
		repo.Branch = strings.TrimPrefix(line, "# branch.head ")
	case strings.HasPrefix(line, "# branch.upstream "):
		repo.Upstream = strings.TrimPrefix(line, "# branch.upstream ")
	case strings.HasPrefix(line, "# branch.ab "):
		parts := strings.Fields(line)
		if len(parts) >= 4 {
			repo.Ahead, _ = strconv.Atoi(strings.TrimPrefix(parts[2], "+"))
			behind, _ := strconv.Atoi(strings.TrimPrefix(parts[3], "-"))
			repo.Behind = behind
		}
	}
}

func parseOrdinaryEntry(repo *Repository, line string) {
	// Format: 1 XY sub mH mI mW hH hI path
	parts := strings.SplitN(line, " ", 9)
	if len(parts) < 9 {
		return
	}

	xy := parts[1]
	path := parts[8]

	if len(xy) < 2 {
		return
	}

	indexStatus := xy[0]
	worktreeStatus := xy[1]

	// Index (staged) changes
	if indexStatus != '.' {
		repo.Staged = append(repo.Staged, FileChange{
			Path:   path,
			Status: charToStatus(indexStatus),
		})
	}

	// Worktree (unstaged) changes
	if worktreeStatus != '.' {
		repo.Unstaged = append(repo.Unstaged, FileChange{
			Path:   path,
			Status: charToStatus(worktreeStatus),
		})
	}
}

func parseRenameEntry(repo *Repository, line string) {
	// Format: 2 XY sub mH mI mW hH hI Xscore path\torigPath
	parts := strings.SplitN(line, " ", 10)
	if len(parts) < 10 {
		return
	}

	xy := parts[1]
	pathPart := parts[9]

	// path and origPath are tab-separated
	paths := strings.SplitN(pathPart, "\t", 2)
	newPath := paths[0]
	oldPath := ""
	if len(paths) > 1 {
		oldPath = paths[1]
	}

	if len(xy) < 2 {
		return
	}

	indexStatus := xy[0]
	worktreeStatus := xy[1]

	if indexStatus != '.' {
		repo.Staged = append(repo.Staged, FileChange{
			Path:    newPath,
			OldPath: oldPath,
			Status:  charToStatus(indexStatus),
		})
	}

	if worktreeStatus != '.' {
		repo.Unstaged = append(repo.Unstaged, FileChange{
			Path:    newPath,
			OldPath: oldPath,
			Status:  charToStatus(worktreeStatus),
		})
	}
}

func parseUnmergedEntry(repo *Repository, line string) {
	// Format: u XY sub m1 m2 m3 mW h1 h2 h3 path
	parts := strings.SplitN(line, " ", 11)
	if len(parts) < 11 {
		return
	}

	path := parts[10]
	repo.Unstaged = append(repo.Unstaged, FileChange{
		Path:   path,
		Status: StatusConflicted,
	})
}

func charToStatus(c byte) FileStatus {
	switch c {
	case 'M':
		return StatusModified
	case 'A':
		return StatusAdded
	case 'D':
		return StatusDeleted
	case 'R':
		return StatusRenamed
	case 'C':
		return StatusCopied
	default:
		return StatusModified
	}
}
