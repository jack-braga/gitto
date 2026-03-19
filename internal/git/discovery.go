package git

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// DiscoverRepos scans the given directory for immediate child directories
// that are git repositories (contain a .git directory or file).
// Returns a sorted list of absolute paths.
func DiscoverRepos(parentDir string) ([]string, error) {
	absDir, err := filepath.Abs(parentDir)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	var repos []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip hidden directories (except we check inside them for .git)
		name := entry.Name()
		if name == "." || name == ".." {
			continue
		}

		childPath := filepath.Join(absDir, name)
		gitPath := filepath.Join(childPath, ".git")

		// .git can be a directory (normal repo) or a file (worktree/submodule)
		if _, err := os.Stat(gitPath); err == nil {
			repos = append(repos, childPath)
		}
	}

	sort.Strings(repos)
	return repos, nil
}
