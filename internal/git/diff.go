package git

import (
	"strconv"
	"strings"
)

// DiffStat holds insertion/deletion counts for a single file.
type DiffStat struct {
	Insertions int
	Deletions  int
}

// GetDiffStats returns per-file insertion/deletion counts.
// If staged is true, shows staged changes (--cached); otherwise unstaged.
func GetDiffStats(repoPath string, staged bool) (map[string]DiffStat, error) {
	args := []string{"diff", "--numstat"}
	if staged {
		args = append(args, "--cached")
	}

	output, err := GitExec(repoPath, args...)
	if err != nil {
		return nil, err
	}

	return parseNumstat(output), nil
}

func parseNumstat(output string) map[string]DiffStat {
	stats := make(map[string]DiffStat)

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}

		// Binary files show "-" for insertions/deletions
		ins, _ := strconv.Atoi(parts[0])
		del, _ := strconv.Atoi(parts[1])
		path := parts[2]

		// Handle renames: "old => new" or "{old => new}/path"
		if idx := strings.Index(path, " => "); idx >= 0 {
			// Use the new path
			path = path[idx+4:]
			path = strings.TrimSuffix(path, "}")
		}

		stats[path] = DiffStat{Insertions: ins, Deletions: del}
	}

	return stats
}
