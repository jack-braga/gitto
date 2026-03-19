package git

import (
	"strings"
	"time"
)

// ListBranches returns all local and remote branches.
func ListBranches(repoPath string) ([]Branch, error) {
	format := "%(refname:short)\t%(objecttype)\t%(HEAD)\t%(upstream:short)\t%(committerdate:iso8601)\t%(subject)"
	output, err := GitExec(repoPath, "branch", "-a", "--format="+format)
	if err != nil {
		return nil, err
	}

	var branches []Branch
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 6)
		if len(parts) < 6 {
			continue
		}

		name := parts[0]
		isCurrent := parts[2] == "*"
		upstream := parts[3]
		dateStr := parts[4]
		message := parts[5]

		isRemote := strings.HasPrefix(name, "remotes/") || strings.Contains(name, "/")
		// Clean up remote branch names
		if strings.HasPrefix(name, "remotes/") {
			name = strings.TrimPrefix(name, "remotes/")
			isRemote = true
		}

		// Skip HEAD pointer entries like "origin/HEAD"
		if strings.HasSuffix(name, "/HEAD") {
			continue
		}

		commitDate, _ := time.Parse("2006-01-02 15:04:05 -0700", dateStr)

		branches = append(branches, Branch{
			Name:        name,
			IsRemote:    isRemote,
			IsCurrent:   isCurrent,
			Upstream:    upstream,
			LastCommit:  commitDate,
			LastMessage: message,
		})
	}

	return branches, nil
}

// SwitchBranch checks out the given branch.
func SwitchBranch(repoPath, branchName string) error {
	_, err := GitExec(repoPath, "checkout", branchName)
	return err
}

// CreateBranch creates a new branch and switches to it.
func CreateBranch(repoPath, branchName string) error {
	_, err := GitExec(repoPath, "checkout", "-b", branchName)
	return err
}

// DeleteBranch deletes a local branch.
func DeleteBranch(repoPath, branchName string) error {
	_, err := GitExec(repoPath, "branch", "-d", branchName)
	return err
}
