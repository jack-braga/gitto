package git

import (
	"fmt"
	"strings"
	"time"
)

const logSeparator = "\x00"

// GetLog returns the most recent commits for a repository.
func GetLog(repoPath string, limit int) ([]LogEntry, error) {
	format := strings.Join([]string{"%h", "%H", "%s", "%an", "%ci", "%D"}, logSeparator)

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
