package git

import (
	"os"
	"path/filepath"
)

// DiscardFile discards changes in a single file.
// For tracked files, uses git checkout. For untracked files, removes the file.
func DiscardFile(repoPath, filePath string, isUntracked bool) error {
	if isUntracked {
		return os.Remove(filepath.Join(repoPath, filePath))
	}
	_, err := GitExec(repoPath, "checkout", "--", filePath)
	return err
}

// DiscardAll discards all changes (tracked and untracked).
func DiscardAll(repoPath string) error {
	if _, err := GitExec(repoPath, "checkout", "--", "."); err != nil {
		return err
	}
	_, err := GitExec(repoPath, "clean", "-fd")
	return err
}
