package git

import (
	"fmt"
	"strings"
	"time"
)

// Use a unit separator (ASCII 31) as delimiter — safe in args unlike NUL.
const logSeparator = "\x1f"

// GetLog returns the most recent commits for a repository.
func GetLog(repoPath string, limit int) ([]LogEntry, error) {
	// %x1f is the unit separator in git format strings
	format := "%h%x1f%H%x1f%s%x1f%an%x1f%ci%x1f%D"

	output, err := GitExec(repoPath, "log",
		"--format="+format,
		fmt.Sprintf("-n%d", limit),
	)
	if err != nil {
		return nil, err
	}

	var entries []LogEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, logSeparator, 6)
		if len(parts) < 6 {
			continue
		}

		date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[4])

		var refs []string
		if parts[5] != "" {
			for _, ref := range strings.Split(parts[5], ", ") {
				refs = append(refs, strings.TrimSpace(ref))
			}
		}

		entries = append(entries, LogEntry{
			Hash:     parts[0],
			FullHash: parts[1],
			Message:  parts[2],
			Author:   parts[3],
			Date:     date,
			Refs:     refs,
		})
	}

	return entries, nil
}
