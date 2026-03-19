package git

// StageFile stages a single file.
func StageFile(repoPath, filePath string) error {
	_, err := GitExec(repoPath, "add", "--", filePath)
	return err
}

// StageAll stages all changes in the repository.
func StageAll(repoPath string) error {
	_, err := GitExec(repoPath, "add", "-A")
	return err
}

// UnstageFile removes a single file from the staging area.
func UnstageFile(repoPath, filePath string) error {
	_, err := GitExec(repoPath, "reset", "HEAD", "--", filePath)
	return err
}

// UnstageAll removes all files from the staging area.
func UnstageAll(repoPath string) error {
	_, err := GitExec(repoPath, "reset", "HEAD")
	return err
}
