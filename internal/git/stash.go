package git

import (
	"fmt"
	"strings"
	"time"
)

// ListStashes returns all stash entries for the repo.
func ListStashes(repoPath string) ([]Stash, error) {
	output, err := GitExec(repoPath, "stash", "list", "--format=%gd\t%s\t%ci")
	if err != nil {
		return nil, err
	}

	var stashes []Stash
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}

		// Parse stash index from "stash@{0}"
		ref := parts[0]
		idx := 0
		if start := strings.Index(ref, "{"); start >= 0 {
			if end := strings.Index(ref, "}"); end > start {
				fmt.Sscanf(ref[start+1:end], "%d", &idx)
			}
		}

		date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[2])

		stashes = append(stashes, Stash{
			Index:   idx,
			Message: parts[1],
			Date:    date,
		})
	}

	return stashes, nil
}

// StashSave creates a new stash with the given message.
func StashSave(repoPath, message string) error {
	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}
	_, err := GitExec(repoPath, args...)
	return err
}

// StashPop pops the stash at the given index.
func StashPop(repoPath string, index int) error {
	_, err := GitExec(repoPath, "stash", "pop", fmt.Sprintf("stash@{%d}", index))
	return err
}

// StashApply applies the stash at the given index without removing it.
func StashApply(repoPath string, index int) error {
	_, err := GitExec(repoPath, "stash", "apply", fmt.Sprintf("stash@{%d}", index))
	return err
}

// StashDrop removes the stash at the given index.
func StashDrop(repoPath string, index int) error {
	_, err := GitExec(repoPath, "stash", "drop", fmt.Sprintf("stash@{%d}", index))
	return err
}
